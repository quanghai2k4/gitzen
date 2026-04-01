// Package updater cung cấp chức năng tự động cập nhật GitZen
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitzen/internal/logger"
)

const (
	// GitHub repository for GitZen
	GitHubRepo = "quanghai2k4/gitzen"
	
	// Timeout for HTTP requests
	HTTPTimeout = 30 * time.Second
	
	// User agent for API requests
	UserAgent = "GitZen-Updater/1.0"
)

// Updater handles the update process for GitZen
type Updater struct {
	currentVersion string
	github         *GitHubClient
	platform       *PlatformDetector
	logger         *logger.Logger
}

// NewUpdater creates a new updater instance
func NewUpdater(currentVersion string) *Updater {
	return &Updater{
		currentVersion: currentVersion,
		github:         NewGitHubClient(GitHubRepo),
		platform:       NewPlatformDetector(),
		logger:         logger.Get(),
	}
}

// UpdateOptions contains options for the update process
type UpdateOptions struct {
	DryRun      bool // Show what would be updated without actually updating
	Force       bool // Force update even if current version is latest
	Backup      bool // Create backup of current binary (default: true)
	Verbose     bool // Show detailed progress information
}

// UpdateResult contains the result of an update operation
type UpdateResult struct {
	Updated         bool   // Whether an update was performed
	CurrentVersion  string // Current version before update
	NewVersion      string // New version after update (if updated)
	BackupPath      string // Path to backup file (if created)
	Message         string // Human-readable result message
}

// CheckForUpdate checks if a newer version is available
func (u *Updater) CheckForUpdate() (*Release, error) {
	u.logger.Info("Checking for updates from GitHub...")
	
	latest, err := u.github.GetLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	
	u.logger.Info("Latest version: %s, current version: %s", latest.TagName, u.currentVersion)
	
	if u.isNewerVersion(latest.TagName, u.currentVersion) {
		return latest, nil
	}
	
	return nil, nil // No update available
}

// Update performs the update process
func (u *Updater) Update(options UpdateOptions) (*UpdateResult, error) {
	result := &UpdateResult{
		CurrentVersion: u.currentVersion,
	}
	
	// Check for updates
	latest, err := u.CheckForUpdate()
	if err != nil {
		return result, err
	}
	
	if latest == nil && !options.Force {
		result.Message = fmt.Sprintf("Already up to date! (v%s)", u.currentVersion)
		return result, nil
	}
	
	if latest == nil && options.Force {
		return result, fmt.Errorf("no releases available to force update to")
	}
	
	result.NewVersion = latest.TagName
	
	if options.DryRun {
		result.Message = fmt.Sprintf("Would update from v%s to v%s", u.currentVersion, latest.TagName)
		return result, nil
	}
	
	if options.Verbose {
		fmt.Printf("Updating from v%s to v%s...\n", u.currentVersion, latest.TagName)
	}
	
	// Find the appropriate asset for current platform
	asset, err := u.findPlatformAsset(latest)
	if err != nil {
		return result, fmt.Errorf("failed to find compatible binary: %w", err)
	}
	
	// Download the new binary
	tempFile, err := u.downloadAsset(asset, options.Verbose)
	if err != nil {
		return result, fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tempFile) // Clean up temp file
	
	// Verify checksum if available
	if err := u.verifyChecksum(tempFile, latest, asset.Name); err != nil {
		return result, fmt.Errorf("checksum verification failed: %w", err)
	}
	
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return result, fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}
	
	// Create backup if requested
	if options.Backup {
		backupPath, err := u.createBackup(realPath)
		if err != nil {
			return result, fmt.Errorf("failed to create backup: %w", err)
		}
		result.BackupPath = backupPath
		
		if options.Verbose {
			fmt.Printf("Created backup: %s\n", backupPath)
		}
	}
	
	// Replace the binary
	if err := u.replaceBinary(tempFile, realPath); err != nil {
		return result, fmt.Errorf("failed to replace binary: %w", err)
	}
	
	result.Updated = true
	result.Message = fmt.Sprintf("✓ Update successful! GitZen is now v%s", latest.TagName)
	
	u.logger.Info("Update completed successfully: v%s -> v%s", u.currentVersion, latest.TagName)
	
	return result, nil
}

// findPlatformAsset finds the appropriate release asset for the current platform
func (u *Updater) findPlatformAsset(release *Release) (*ReleaseAsset, error) {
	platform := u.platform.Detect()
	
	u.logger.Debug("Looking for asset matching platform: %s", platform.String())
	
	for _, asset := range release.Assets {
		if u.platform.MatchesAsset(asset.Name, platform) {
			u.logger.Debug("Found matching asset: %s", asset.Name)
			return &asset, nil
		}
	}
	
	return nil, fmt.Errorf("no compatible binary found for %s", platform.String())
}

// downloadAsset downloads a release asset to a temporary file
func (u *Updater) downloadAsset(asset *ReleaseAsset, verbose bool) (string, error) {
	if verbose {
		fmt.Printf("Downloading %s...\n", asset.Name)
	}
	
	u.logger.Info("Downloading asset: %s (%d bytes)", asset.Name, asset.Size)
	
	client := &http.Client{Timeout: HTTPTimeout}
	
	req, err := http.NewRequest("GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("User-Agent", UserAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", "gitzen-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	
	// Copy with progress if verbose
	var written int64
	if verbose && asset.Size > 0 {
		written, err = u.copyWithProgress(tempFile, resp.Body, asset.Size)
	} else {
		written, err = io.Copy(tempFile, resp.Body)
	}
	
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}
	
	u.logger.Debug("Downloaded %d bytes to %s", written, tempFile.Name())
	
	// Make executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}
	
	return tempFile.Name(), nil
}

// copyWithProgress copies data while showing progress
func (u *Updater) copyWithProgress(dst io.Writer, src io.Reader, total int64) (int64, error) {
	var written int64
	buffer := make([]byte, 32*1024)
	
	for {
		n, err := src.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return written, writeErr
			}
			written += int64(n)
			
			// Show progress
			percentage := float64(written) / float64(total) * 100
			fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", percentage, written, total)
		}
		
		if err == io.EOF {
			fmt.Println() // New line after progress
			break
		}
		if err != nil {
			return written, err
		}
	}
	
	return written, nil
}

// verifyChecksum verifies the downloaded file against checksums from the release
func (u *Updater) verifyChecksum(filePath string, release *Release, assetName string) error {
	// Look for checksums file in release assets
	var checksumsAsset *ReleaseAsset
	for _, asset := range release.Assets {
		if strings.Contains(strings.ToLower(asset.Name), "checksum") ||
		   strings.Contains(strings.ToLower(asset.Name), "sha256") {
			checksumsAsset = &asset
			break
		}
	}
	
	if checksumsAsset == nil {
		u.logger.Warn("No checksums file found in release, skipping verification")
		return nil
	}
	
	u.logger.Debug("Downloading checksums from: %s", checksumsAsset.Name)
	
	// Download checksums file
	client := &http.Client{Timeout: HTTPTimeout}
	resp, err := client.Get(checksumsAsset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()
	
	checksumsData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksums: %w", err)
	}
	
	// Parse checksums and find the one for our asset
	expectedChecksum, err := u.parseChecksum(string(checksumsData), assetName)
	if err != nil {
		return fmt.Errorf("failed to parse checksum for %s: %w", assetName, err)
	}
	
	// Calculate actual checksum
	actualChecksum, err := u.calculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}
	
	u.logger.Info("Checksum verification passed")
	return nil
}

// parseChecksum parses a checksums file and returns the checksum for the specified file
func (u *Updater) parseChecksum(checksumsText, filename string) (string, error) {
	lines := strings.Split(checksumsText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Handle both "checksum filename" and "checksum  filename" formats
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			file := strings.Join(parts[1:], " ")
			
			// Remove any path separators from filename comparison
			baseName := filepath.Base(file)
			targetName := filepath.Base(filename)
			
			if baseName == targetName {
				return checksum, nil
			}
		}
	}
	
	return "", fmt.Errorf("checksum not found for file: %s", filename)
}

// calculateSHA256 calculates the SHA256 checksum of a file
func (u *Updater) calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// createBackup creates a backup of the current binary
func (u *Updater) createBackup(execPath string) (string, error) {
	backupPath := execPath + ".backup." + u.currentVersion
	
	// Remove existing backup if it exists
	os.Remove(backupPath)
	
	// Copy current binary to backup location
	src, err := os.Open(execPath)
	if err != nil {
		return "", err
	}
	defer src.Close()
	
	dst, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(backupPath)
		return "", err
	}
	
	// Copy permissions
	srcInfo, err := src.Stat()
	if err != nil {
		return backupPath, nil // Return backup path even if we can't copy permissions
	}
	
	os.Chmod(backupPath, srcInfo.Mode())
	
	return backupPath, nil
}

// replaceBinary replaces the current binary with the new one
func (u *Updater) replaceBinary(newBinaryPath, targetPath string) error {
	// On Windows, we might need special handling for replacing a running executable
	if runtime.GOOS == "windows" {
		return u.replaceWindowsBinary(newBinaryPath, targetPath)
	}
	
	// Copy new binary over the old one
	src, err := os.Open(newBinaryPath)
	if err != nil {
		return err
	}
	defer src.Close()
	
	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	
	// Ensure executable permissions
	return os.Chmod(targetPath, 0755)
}

// replaceWindowsBinary handles binary replacement on Windows
func (u *Updater) replaceWindowsBinary(newBinaryPath, targetPath string) error {
	// On Windows, we can't replace a running executable directly
	// Move current binary to .old and move new binary to target location
	oldPath := targetPath + ".old"
	
	// Remove any existing .old file
	os.Remove(oldPath)
	
	// Move current binary to .old
	if err := os.Rename(targetPath, oldPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}
	
	// Move new binary to target location
	if err := os.Rename(newBinaryPath, targetPath); err != nil {
		// Try to restore original if move failed
		os.Rename(oldPath, targetPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}
	
	// Schedule .old file for deletion on next boot (Windows-specific)
	// This is optional and we ignore errors
	if err := u.scheduleFileForDeletion(oldPath); err != nil {
		u.logger.Warn("Failed to schedule old binary for deletion: %v", err)
	}
	
	return nil
}

// scheduleFileForDeletion schedules a file for deletion on Windows
func (u *Updater) scheduleFileForDeletion(filePath string) error {
	// This would use Windows API calls to schedule file deletion
	// For now, we just try to delete it immediately
	return os.Remove(filePath)
}

// isNewerVersion compares two version strings and returns true if newVersion is newer
func (u *Updater) isNewerVersion(newVersion, currentVersion string) bool {
	// Remove 'v' prefix if present
	newVersion = strings.TrimPrefix(newVersion, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	
	// Handle dev/unknown versions
	if currentVersion == "dev" || currentVersion == "unknown" || currentVersion == "" {
		return true
	}
	
	// Simple string comparison for now
	// In a production system, you'd want proper semantic version parsing
	return newVersion != currentVersion
}