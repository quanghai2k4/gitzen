package updater

import (
	"runtime"
	"strings"

	"gitzen/internal/logger"
)

// PlatformDetector handles platform detection and asset matching
type PlatformDetector struct {
	logger *logger.Logger
}

// Platform represents the current system platform
type Platform struct {
	OS   string // linux, windows, darwin
	Arch string // amd64, arm64, 386
}

// NewPlatformDetector creates a new platform detector
func NewPlatformDetector() *PlatformDetector {
	return &PlatformDetector{
		logger: logger.Get(),
	}
}

// Detect detects the current platform
func (p *PlatformDetector) Detect() Platform {
	platform := Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	p.logger.Debug("Detected platform: %s", platform.String())

	return platform
}

// String returns a string representation of the platform
func (p Platform) String() string {
	return p.OS + "/" + p.Arch
}

// MatchesAsset checks if an asset filename matches the current platform
func (p *PlatformDetector) MatchesAsset(assetName string, platform Platform) bool {
	assetLower := strings.ToLower(assetName)

	// Skip checksums files
	if strings.Contains(assetLower, "checksum") ||
		strings.Contains(assetLower, "sha256") ||
		strings.HasSuffix(assetLower, ".txt") {
		return false
	}

	// Map Go arch names to common binary naming conventions
	archPatterns := p.getArchPatterns(platform.Arch)
	osPatterns := p.getOSPatterns(platform.OS)

	// Check if asset contains OS pattern
	osMatch := false
	for _, osPattern := range osPatterns {
		if strings.Contains(assetLower, osPattern) {
			osMatch = true
			break
		}
	}

	if !osMatch {
		return false
	}

	// Check if asset contains arch pattern
	archMatch := false
	for _, archPattern := range archPatterns {
		if strings.Contains(assetLower, archPattern) {
			archMatch = true
			break
		}
	}

	if !archMatch {
		return false
	}

	// Additional validation - ensure it's an executable
	return p.isExecutableAsset(assetName, platform.OS)
}

// getOSPatterns returns possible naming patterns for an OS
func (p *PlatformDetector) getOSPatterns(os string) []string {
	switch os {
	case "linux":
		return []string{"linux"}
	case "darwin":
		return []string{"darwin", "macos", "mac"}
	case "windows":
		return []string{"windows", "win"}
	default:
		return []string{os}
	}
}

// getArchPatterns returns possible naming patterns for an architecture
func (p *PlatformDetector) getArchPatterns(arch string) []string {
	switch arch {
	case "amd64":
		return []string{"amd64", "x86_64", "x64"}
	case "arm64":
		return []string{"arm64", "aarch64"}
	case "386":
		return []string{"386", "i386", "x86"}
	case "arm":
		return []string{"arm", "armv6", "armv7"}
	default:
		return []string{arch}
	}
}

// isExecutableAsset checks if an asset is likely an executable for the given OS
func (p *PlatformDetector) isExecutableAsset(assetName, os string) bool {
	assetLower := strings.ToLower(assetName)

	switch os {
	case "windows":
		// Windows executables should have .exe extension
		return strings.HasSuffix(assetLower, ".exe")
	case "linux", "darwin":
		// Unix-like systems typically don't use extensions for executables
		// Exclude common archive formats
		excludeExtensions := []string{
			".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tar.xz", ".txz",
			".zip", ".7z", ".rar",
			".deb", ".rpm", ".pkg", ".dmg",
			".txt", ".md", ".json", ".xml", ".yaml", ".yml",
			".sig", ".asc", // signature files
		}

		for _, ext := range excludeExtensions {
			if strings.HasSuffix(assetLower, ext) {
				return false
			}
		}

		return true
	default:
		return true
	}
}

// GetExpectedAssetName returns the expected asset name pattern for a platform
func (p *PlatformDetector) GetExpectedAssetName(binaryName string, platform Platform) []string {
	var patterns []string

	osPatterns := p.getOSPatterns(platform.OS)
	archPatterns := p.getArchPatterns(platform.Arch)

	// Generate common naming patterns
	for _, osPattern := range osPatterns {
		for _, archPattern := range archPatterns {
			// Common patterns used by GoReleaser and similar tools
			patterns = append(patterns, []string{
				binaryName + "_" + osPattern + "_" + archPattern,
				binaryName + "-" + osPattern + "-" + archPattern,
				binaryName + "_" + platform.OS + "_" + platform.Arch,
				binaryName + "-" + platform.OS + "-" + platform.Arch,
			}...)

			// With extension for Windows
			if platform.OS == "windows" {
				patterns = append(patterns, []string{
					binaryName + "_" + osPattern + "_" + archPattern + ".exe",
					binaryName + "-" + osPattern + "-" + archPattern + ".exe",
					binaryName + "_" + platform.OS + "_" + platform.Arch + ".exe",
					binaryName + "-" + platform.OS + "-" + platform.Arch + ".exe",
				}...)
			}
		}
	}

	return patterns
}

// SupportedPlatforms returns a list of commonly supported platforms
func (p *PlatformDetector) SupportedPlatforms() []Platform {
	return []Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "linux", Arch: "386"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "windows", Arch: "amd64"},
		{OS: "windows", Arch: "386"},
	}
}

// IsSupportedPlatform checks if a platform is commonly supported
func (p *PlatformDetector) IsSupportedPlatform(platform Platform) bool {
	supported := p.SupportedPlatforms()
	for _, sp := range supported {
		if sp.OS == platform.OS && sp.Arch == platform.Arch {
			return true
		}
	}
	return false
}
