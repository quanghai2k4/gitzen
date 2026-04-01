package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadRepoConfig tải cấu hình từ .git/gitzen-config.yml, trả về defaults nếu file không tồn tại
func LoadRepoConfig(repoRoot string) (*RepoConfig, error) {
	configPath := filepath.Join(repoRoot, ".git", "gitzen-config.yml")

	// Nếu file không tồn tại, trả về default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return NewDefaultConfig(), nil
	}

	// Đọc file YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		return NewDefaultConfig(), fmt.Errorf("cannot read config file %s: %w", configPath, err)
	}

	var config RepoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return NewDefaultConfig(), fmt.Errorf("cannot parse YAML config %s: %w", configPath, err)
	}

	// Validation check
	if !config.IsValid() {
		return NewDefaultConfig(), fmt.Errorf("invalid configuration in %s, using defaults", configPath)
	}

	return &config, nil
}

// SaveRepoConfig lưu cấu hình vào .git/gitzen-config.yml
func SaveRepoConfig(repoRoot string, config *RepoConfig) error {
	configPath := filepath.Join(repoRoot, ".git", "gitzen-config.yml")

	// Ensure .git directory exists
	gitDir := filepath.Join(repoRoot, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return fmt.Errorf("cannot create .git directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("cannot marshal config to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write config file %s: %w", configPath, err)
	}

	return nil
}

// NewDefaultConfig tạo cấu hình mặc định theo nghiên cứu
func NewDefaultConfig() *RepoConfig {
	return &RepoConfig{
		AutoFetch: AutoFetchConfig{
			Enabled:         true,
			StartupFetch:    true,
			TargetBranches:  []string{"auto"}, // "auto" nghĩa là main + current branch
			IntervalMinutes: 30,
		},
		FileWatch: FileWatchConfig{
			Enabled:     true,
			DebounceMs:  300,
			IgnoredDirs: []string{"node_modules", "vendor", ".next", "dist", "build", ".cache", ".tmp", "__pycache__", ".opencode"},
		},
	}
}
