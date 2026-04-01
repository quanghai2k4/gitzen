package background

import (
	tea "github.com/charmbracelet/bubbletea"

	"gitzen/internal/config"
	"gitzen/internal/logger"
)

// ExecuteAutoFetch coordina background fetch với configuration
func (m *Manager) ExecuteAutoFetch(repoRoot string) tea.Cmd {
	return func() tea.Msg {
		log := logger.Get()

		// Load repository configuration
		repoConfig, err := config.LoadRepoConfig(repoRoot)
		if err != nil {
			log.Warn("auto fetch: cannot load config: %v", err)
			// Continue with default config if loading fails
			repoConfig = config.NewDefaultConfig()
		}

		// Skip execution if AutoFetch is disabled
		if !repoConfig.AutoFetch.Enabled {
			log.Debug("auto fetch: disabled in configuration")
			return autoFetchResultMsg{Success: true, Skipped: true, Message: "auto fetch disabled"}
		}

		// Determine target branches
		var targetBranches []string
		if len(repoConfig.AutoFetch.TargetBranches) > 0 && repoConfig.AutoFetch.TargetBranches[0] == "auto" {
			// Auto mode: fetch main + current branch
			defaultBranch, err := m.gitRunner.GetDefaultBranch("origin")
			if err != nil {
				log.Warn("auto fetch: cannot get default branch: %v", err)
				defaultBranch = "main" // fallback
			}

			currentBranch, err := m.gitRunner.GetCurrentBranch()
			if err != nil {
				log.Warn("auto fetch: cannot get current branch: %v", err)
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

		// Use Manager.ExecuteIfSafe() to ensure working directory safety
		err = m.ExecuteIfSafe(func() error {
			log.Info("auto fetch: fetching branches %v", targetBranches)
			return m.gitRunner.FetchBranches("origin", targetBranches)
		})

		if err != nil {
			log.Warn("auto fetch: failed or skipped: %v", err)
			return autoFetchResultMsg{Success: false, Message: err.Error()}
		}

		log.Info("auto fetch: completed successfully for branches %v", targetBranches)
		return autoFetchResultMsg{Success: true, Message: "auto fetch completed"}
	}
}

// autoFetchResultMsg reports background fetch results
type autoFetchResultMsg struct {
	Success bool
	Skipped bool
	Message string
}
