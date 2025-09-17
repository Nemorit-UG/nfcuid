package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// UpdateChecker handles checking for and installing updates
type UpdateChecker struct {
	config              *Config
	notificationManager *NotificationManager
	currentVersion      string
	githubOwner         string
	githubRepo          string
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker(config *Config, notificationManager *NotificationManager) *UpdateChecker {
	return &UpdateChecker{
		config:              config,
		notificationManager: notificationManager,
		currentVersion:      Version,
		githubOwner:         GitHubOwner,
		githubRepo:          GitHubRepo,
	}
}

// CheckForUpdates checks if a newer version is available
func (uc *UpdateChecker) CheckForUpdates() (*GitHubRelease, bool, error) {
	if !uc.config.Updates.Enabled {
		return nil, false, nil
	}

	fmt.Println("Checking for updates...")

	// Get latest release from GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", uc.githubOwner, uc.githubRepo)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to parse release data: %v", err)
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		fmt.Printf("Latest release %s is draft/prerelease, skipping\n", release.TagName)
		return nil, false, nil
	}

	// Compare versions
	hasUpdate, err := uc.isNewerVersion(release.TagName, uc.currentVersion)
	if err != nil {
		return nil, false, fmt.Errorf("failed to compare versions: %v", err)
	}

	return &release, hasUpdate, nil
}

// isNewerVersion compares version strings (basic semantic version comparison)
func (uc *UpdateChecker) isNewerVersion(remote, current string) (bool, error) {
	// Remove 'v' prefix if present
	remote = strings.TrimPrefix(remote, "v")
	current = strings.TrimPrefix(current, "v")

	// Split versions into parts
	remoteParts := strings.Split(remote, ".")
	currentParts := strings.Split(current, ".")

	// Ensure both have at least 3 parts (major.minor.patch)
	for len(remoteParts) < 3 {
		remoteParts = append(remoteParts, "0")
	}
	for len(currentParts) < 3 {
		currentParts = append(currentParts, "0")
	}

	// Compare each part
	for i := 0; i < 3; i++ {
		remoteNum, err := strconv.Atoi(remoteParts[i])
		if err != nil {
			return false, fmt.Errorf("invalid remote version part: %s", remoteParts[i])
		}

		currentNum, err := strconv.Atoi(currentParts[i])
		if err != nil {
			return false, fmt.Errorf("invalid current version part: %s", currentParts[i])
		}

		if remoteNum > currentNum {
			return true, nil
		} else if remoteNum < currentNum {
			return false, nil
		}
		// If equal, continue to next part
	}

	return false, nil // Versions are equal
}

// DownloadUpdate downloads the update package for the current platform
func (uc *UpdateChecker) DownloadUpdate(release *GitHubRelease) (string, error) {
	if !uc.config.Updates.AutoDownload {
		return "", fmt.Errorf("auto-download is disabled")
	}

	// Find the appropriate asset for current platform
	assetName := uc.getAssetNameForPlatform(release.TagName)
	var downloadURL string
	var assetSize int64

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			assetSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no suitable asset found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading update: %s (%d bytes)\n", assetName, assetSize)

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "nfcuid-update-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	downloadPath := filepath.Join(tempDir, assetName)

	// Download the file
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create the file
	file, err := os.Create(downloadPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create download file: %v", err)
	}
	defer file.Close()

	// Copy with progress (simplified)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to save download: %v", err)
	}

	fmt.Printf("Update downloaded successfully: %s\n", downloadPath)
	return downloadPath, nil
}

// getAssetNameForPlatform returns the expected asset name for the current platform
func (uc *UpdateChecker) getAssetNameForPlatform(version string) string {
	// Remove 'v' prefix from version
	version = strings.TrimPrefix(version, "v")

	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("nfcuid_windows_amd64_%s.zip", version)
	case "linux":
		return fmt.Sprintf("nfcuid_linux_amd64_%s.tar.gz", version)
	case "darwin":
		return fmt.Sprintf("nfcuid_darwin_amd64_%s.tar.gz", version)
	default:
		return fmt.Sprintf("nfcuid_%s_amd64_%s.tar.gz", runtime.GOOS, version)
	}
}

// InstallUpdate extracts and installs the downloaded update
func (uc *UpdateChecker) InstallUpdate(downloadPath string) error {
	if !uc.config.Updates.AutoInstall {
		fmt.Println("Auto-install is disabled. Update downloaded but not installed.")
		if uc.notificationManager != nil {
			uc.notificationManager.NotifyInfo("Update Available", fmt.Sprintf("Update downloaded to %s. Manual installation required.", downloadPath))
		}
		return nil
	}

	fmt.Println("Installing update...")

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	// Create temp directory for extraction
	extractDir, err := os.MkdirTemp("", "nfcuid-extract-")
	if err != nil {
		return fmt.Errorf("failed to create extraction directory: %v", err)
	}
	defer os.RemoveAll(extractDir)

	// Extract the downloaded file
	var newExePath string
	if strings.HasSuffix(downloadPath, ".zip") {
		newExePath, err = uc.extractZip(downloadPath, extractDir)
	} else {
		return fmt.Errorf("unsupported archive format, only .zip is currently supported")
	}

	if err != nil {
		return fmt.Errorf("failed to extract update: %v", err)
	}

	// Backup current executable
	backupPath := currentExe + ".backup"
	if err := copyFile(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %v", err)
	}

	// Replace current executable
	if err := copyFile(newExePath, currentExe); err != nil {
		// Restore backup on failure
		copyFile(backupPath, currentExe)
		return fmt.Errorf("failed to replace executable: %v", err)
	}

	// Make executable (Unix systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(currentExe, 0755); err != nil {
			fmt.Printf("Warning: failed to set executable permissions: %v\n", err)
		}
	}

	fmt.Println("Update installed successfully!")

	if uc.notificationManager != nil {
		uc.notificationManager.NotifyInfo("Update Installed", "Application has been updated. Restart to use the new version.")
	}

	// Clean up backup after successful installation
	os.Remove(backupPath)

	return nil
}

// extractZip extracts a ZIP file and returns the path to the executable
func (uc *UpdateChecker) extractZip(zipPath, extractDir string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var executablePath string
	executableName := "nfcuid"
	if runtime.GOOS == "windows" {
		executableName = "nfcuid.exe"
	}

	for _, file := range reader.File {
		// Extract only the executable file
		if strings.HasSuffix(file.Name, executableName) {
			extractPath := filepath.Join(extractDir, filepath.Base(file.Name))

			rc, err := file.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			outFile, err := os.Create(extractPath)
			if err != nil {
				return "", err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return "", err
			}

			executablePath = extractPath
			break
		}
	}

	if executablePath == "" {
		return "", fmt.Errorf("executable not found in archive")
	}

	return executablePath, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// PerformUpdateCheck performs a complete update check and installation if configured
func (uc *UpdateChecker) PerformUpdateCheck() error {
	release, hasUpdate, err := uc.CheckForUpdates()
	if err != nil {
		fmt.Printf("Failed to check for updates: %v\n", err)
		if uc.notificationManager != nil {
			uc.notificationManager.NotifyErrorThrottled("update-check-error", fmt.Sprintf("Failed to check for updates: %v", err))
		}
		return err
	}

	if !hasUpdate {
		fmt.Println("No updates available")
		return nil
	}

	fmt.Printf("Update available: %s -> %s\n", uc.currentVersion, release.TagName)
	if uc.notificationManager != nil {
		uc.notificationManager.NotifyInfo("Update Available", fmt.Sprintf("New version %s is available", release.TagName))
	}

	if !uc.config.Updates.AutoDownload {
		fmt.Println("Auto-download is disabled. Please download manually from GitHub releases.")
		return nil
	}

	downloadPath, err := uc.DownloadUpdate(release)
	if err != nil {
		fmt.Printf("Failed to download update: %v\n", err)
		if uc.notificationManager != nil {
			uc.notificationManager.NotifyErrorThrottled("update-download-error", fmt.Sprintf("Failed to download update: %v", err))
		}
		return err
	}

	if uc.config.Updates.AutoInstall {
		err = uc.InstallUpdate(downloadPath)
		if err != nil {
			fmt.Printf("Failed to install update: %v\n", err)
			if uc.notificationManager != nil {
				uc.notificationManager.NotifyErrorThrottled("update-install-error", fmt.Sprintf("Failed to install update: %v", err))
			}
			return err
		}
	} else {
		fmt.Printf("Update downloaded to: %s\n", downloadPath)
		if uc.notificationManager != nil {
			uc.notificationManager.NotifyInfo("Update Downloaded", fmt.Sprintf("Update downloaded to %s. Set auto_install to true to install automatically.", downloadPath))
		}
	}

	return nil
}
