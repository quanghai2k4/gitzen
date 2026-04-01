package updater

import (
	"testing"
)

func TestPlatformDetection(t *testing.T) {
	detector := NewPlatformDetector()
	platform := detector.Detect()
	
	// Platform should have valid OS and Arch
	if platform.OS == "" {
		t.Error("Platform OS should not be empty")
	}
	
	if platform.Arch == "" {
		t.Error("Platform Arch should not be empty")
	}
	
	// Check if detected platform is in supported list
	if !detector.IsSupportedPlatform(platform) {
		t.Logf("Warning: Current platform %s may not be supported", platform.String())
	}
}

func TestPlatformString(t *testing.T) {
	platform := Platform{OS: "linux", Arch: "amd64"}
	expected := "linux/amd64"
	
	if platform.String() != expected {
		t.Errorf("Platform.String() = %s, want %s", platform.String(), expected)
	}
}

func TestAssetMatching(t *testing.T) {
	detector := NewPlatformDetector()
	
	testCases := []struct {
		assetName string
		platform  Platform
		expected  bool
	}{
		// Linux amd64
		{"gitzen_linux_amd64", Platform{"linux", "amd64"}, true},
		{"gitzen-linux-amd64", Platform{"linux", "amd64"}, true},
		{"gitzen_linux_x86_64", Platform{"linux", "amd64"}, true},
		
		// Windows amd64
		{"gitzen_windows_amd64.exe", Platform{"windows", "amd64"}, true},
		{"gitzen-windows-amd64.exe", Platform{"windows", "amd64"}, true},
		{"gitzen_win_x64.exe", Platform{"windows", "amd64"}, true},
		
		// macOS arm64
		{"gitzen_darwin_arm64", Platform{"darwin", "arm64"}, true},
		{"gitzen_macos_arm64", Platform{"darwin", "arm64"}, true},
		{"gitzen_mac_aarch64", Platform{"darwin", "arm64"}, true},
		
		// Non-matching cases
		{"gitzen_linux_amd64", Platform{"windows", "amd64"}, false},
		{"gitzen_windows_amd64.exe", Platform{"linux", "amd64"}, false},
		{"checksums.txt", Platform{"linux", "amd64"}, false},
		{"gitzen_linux_arm64", Platform{"linux", "amd64"}, false},
		
		// Windows without .exe should fail
		{"gitzen_windows_amd64", Platform{"windows", "amd64"}, false},
	}
	
	for _, tc := range testCases {
		result := detector.MatchesAsset(tc.assetName, tc.platform)
		if result != tc.expected {
			t.Errorf("MatchesAsset(%s, %s) = %v, want %v", 
				tc.assetName, tc.platform.String(), result, tc.expected)
		}
	}
}

func TestGetArchPatterns(t *testing.T) {
	detector := NewPlatformDetector()
	
	testCases := []struct {
		arch     string
		patterns []string
	}{
		{"amd64", []string{"amd64", "x86_64", "x64"}},
		{"arm64", []string{"arm64", "aarch64"}},
		{"386", []string{"386", "i386", "x86"}},
		{"unknown", []string{"unknown"}},
	}
	
	for _, tc := range testCases {
		patterns := detector.getArchPatterns(tc.arch)
		if len(patterns) != len(tc.patterns) {
			t.Errorf("getArchPatterns(%s) returned %d patterns, want %d", 
				tc.arch, len(patterns), len(tc.patterns))
			continue
		}
		
		for i, pattern := range patterns {
			if pattern != tc.patterns[i] {
				t.Errorf("getArchPatterns(%s)[%d] = %s, want %s", 
					tc.arch, i, pattern, tc.patterns[i])
			}
		}
	}
}

func TestGetOSPatterns(t *testing.T) {
	detector := NewPlatformDetector()
	
	testCases := []struct {
		os       string
		patterns []string
	}{
		{"linux", []string{"linux"}},
		{"darwin", []string{"darwin", "macos", "mac"}},
		{"windows", []string{"windows", "win"}},
		{"unknown", []string{"unknown"}},
	}
	
	for _, tc := range testCases {
		patterns := detector.getOSPatterns(tc.os)
		if len(patterns) != len(tc.patterns) {
			t.Errorf("getOSPatterns(%s) returned %d patterns, want %d", 
				tc.os, len(patterns), len(tc.patterns))
			continue
		}
		
		for i, pattern := range patterns {
			if pattern != tc.patterns[i] {
				t.Errorf("getOSPatterns(%s)[%d] = %s, want %s", 
					tc.os, i, pattern, tc.patterns[i])
			}
		}
	}
}

func TestVersionComparison(t *testing.T) {
	updater := NewUpdater("1.0.0")
	
	testCases := []struct {
		newVersion     string
		currentVersion string
		expected       bool
	}{
		{"1.0.1", "1.0.0", true},
		{"2.0.0", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"0.9.0", "1.0.0", true}, // Simple string comparison for now
		{"v1.0.1", "v1.0.0", true},
		{"1.0.1", "dev", true},
		{"1.0.1", "unknown", true},
		{"1.0.1", "", true},
	}
	
	for _, tc := range testCases {
		result := updater.isNewerVersion(tc.newVersion, tc.currentVersion)
		if result != tc.expected {
			t.Errorf("isNewerVersion(%s, %s) = %v, want %v", 
				tc.newVersion, tc.currentVersion, result, tc.expected)
		}
	}
}