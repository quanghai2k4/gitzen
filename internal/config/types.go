package config

// RepoConfig đại diện cho cấu hình của một repository cụ thể
type RepoConfig struct {
	AutoFetch AutoFetchConfig `yaml:"auto_fetch"`
	FileWatch FileWatchConfig `yaml:"file_watch"`
}

// AutoFetchConfig chứa các cài đặt cho tính năng auto fetch
type AutoFetchConfig struct {
	Enabled         bool     `yaml:"enabled"`
	StartupFetch    bool     `yaml:"startup_fetch"`
	TargetBranches  []string `yaml:"target_branches"`
	IntervalMinutes int      `yaml:"interval_minutes"`
}

// FileWatchConfig chứa các cài đặt cho tính năng file watching
type FileWatchConfig struct {
	Enabled       bool     `yaml:"enabled"`
	DebounceMs    int      `yaml:"debounce_ms"`
	IgnoredDirs   []string `yaml:"ignored_dirs"`
}

// IsValid kiểm tra tính hợp lệ của cấu hình
func (c *RepoConfig) IsValid() bool {
	// IntervalMinutes phải dương
	if c.AutoFetch.IntervalMinutes <= 0 {
		return false
	}

	// TargetBranches không được empty
	if len(c.AutoFetch.TargetBranches) == 0 {
		return false
	}

	return true
}