package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRepoConfig(t *testing.T) {
	// Create temporary directory for test repo
	tmpDir, err := os.MkdirTemp("", "gitzen-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Test 1: LoadRepoConfig trả về default config khi file không tồn tại
	config, err := LoadRepoConfig(tmpDir)
	if err != nil {
		t.Errorf("LoadRepoConfig should not error with missing file, got: %v", err)
	}
	if config == nil {
		t.Fatal("LoadRepoConfig should return non-nil config")
	}
	if !config.AutoFetch.Enabled {
		t.Error("Default config should have AutoFetch.Enabled = true")
	}
	if !config.AutoFetch.StartupFetch {
		t.Error("Default config should have AutoFetch.StartupFetch = true")
	}
	if config.AutoFetch.IntervalMinutes != 30 {
		t.Errorf("Default config should have AutoFetch.IntervalMinutes = 30, got: %d", config.AutoFetch.IntervalMinutes)
	}

	// Test 2: SaveRepoConfig tạo YAML file hợp lệ
	err = SaveRepoConfig(tmpDir, config)
	if err != nil {
		t.Errorf("SaveRepoConfig should not error, got: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(gitDir, "gitzen-config.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("SaveRepoConfig should create gitzen-config.yml file")
	}

	// Test 3: LoadRepoConfig đọc YAML file đã tạo
	config2, err := LoadRepoConfig(tmpDir)
	if err != nil {
		t.Errorf("LoadRepoConfig should read saved file without error, got: %v", err)
	}
	if config2.AutoFetch.Enabled != config.AutoFetch.Enabled {
		t.Error("Loaded config should match saved config")
	}
}

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()
	if config == nil {
		t.Fatal("NewDefaultConfig should return non-nil config")
	}

	// Verify default values per research
	if !config.AutoFetch.Enabled {
		t.Error("Default config should have Enabled = true")
	}
	if !config.AutoFetch.StartupFetch {
		t.Error("Default config should have StartupFetch = true")  
	}
	if config.AutoFetch.IntervalMinutes != 30 {
		t.Errorf("Default config should have IntervalMinutes = 30, got: %d", config.AutoFetch.IntervalMinutes)
	}
	if len(config.AutoFetch.TargetBranches) != 1 || config.AutoFetch.TargetBranches[0] != "auto" {
		t.Errorf("Default config should have TargetBranches = [\"auto\"], got: %v", config.AutoFetch.TargetBranches)
	}
}

func TestConfigValidation(t *testing.T) {
	config := NewDefaultConfig()
	
	// Test valid config
	if !config.IsValid() {
		t.Error("Default config should be valid")
	}

	// Test invalid config: negative interval
	config.AutoFetch.IntervalMinutes = -5
	if config.IsValid() {
		t.Error("Config with negative interval should be invalid")
	}
}