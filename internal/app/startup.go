package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/config"
	"gitzen/internal/logger"
)

// startupFetchCmd khởi tạo startup fetch trong background
func startupFetchCmd() tea.Cmd {
	return func() tea.Msg {
		return startupFetchMsg{}
	}
}

// handleStartupFetch phối hợp config loading và fetch execution
func handleStartupFetch(m model) tea.Cmd {
	return func() tea.Msg {
		log := logger.Get()

		// Load repository configuration
		repoConfig, err := config.LoadRepoConfig(m.repoRoot)
		if err != nil {
			log.Warn("startup fetch: cannot load config: %v", err)
			// Continue with default config if loading fails
			repoConfig = config.NewDefaultConfig()
		}

		// Check if startup fetch is enabled
		if !repoConfig.AutoFetch.Enabled || !repoConfig.AutoFetch.StartupFetch {
			log.Debug("startup fetch: disabled in configuration")
			return startupFetchResultMsg{Success: true, Skipped: true, Message: "startup fetch disabled"}
		}

		// Determine target branches
		var targetBranches []string
		if len(repoConfig.AutoFetch.TargetBranches) > 0 && repoConfig.AutoFetch.TargetBranches[0] == "auto" {
			// Auto mode: fetch main + current branch
			defaultBranch, err := m.git.GetDefaultBranch("origin")
			if err != nil {
				log.Warn("startup fetch: cannot get default branch: %v", err)
				defaultBranch = "main" // fallback
			}

			currentBranch, err := m.git.GetCurrentBranch()
			if err != nil {
				log.Warn("startup fetch: cannot get current branch: %v", err)
				currentBranch = "HEAD" // fallback
			}

			// Deduplicate branches
			branchSet := make(map[string]bool)
			branchSet[defaultBranch] = true
			if currentBranch != "HEAD" {
				branchSet[currentBranch] = true
			}

			for branch := range branchSet {
				targetBranches = append(targetBranches, branch)
			}
		} else {
			targetBranches = repoConfig.AutoFetch.TargetBranches
		}

		// Execute fetch via background manager for safety
		err = m.backgroundManager.ExecuteIfSafe(func() error {
			log.Info("startup fetch: fetching branches %v", targetBranches)
			return m.git.FetchBranches("origin", targetBranches)
		})

		if err != nil {
			log.Warn("startup fetch: failed: %v", err)
			return startupFetchResultMsg{Success: false, Message: err.Error()}
		}

		log.Info("startup fetch: completed successfully for branches %v", targetBranches)
		return startupFetchResultMsg{Success: true, Message: "startup fetch completed"}
	}
}
