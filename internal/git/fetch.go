package git

import (
	"fmt"
	"strings"
)

// FetchBranches thực hiện git fetch cho các branch cụ thể từ remote
func (r Runner) FetchBranches(remote string, branches []string) error {
	if len(branches) == 0 {
		return nil // No-op cho empty branch list
	}

	// Tạo refspecs cho từng branch: branch:refs/remotes/remote/branch
	var refspecs []string
	for _, branch := range branches {
		refspec := fmt.Sprintf("%s:refs/remotes/%s/%s", branch, remote, branch)
		refspecs = append(refspecs, refspec)
	}

	// Thực hiện git fetch với NetworkTimeout cho network operation
	args := append([]string{"fetch", remote}, refspecs...)
	_, err := r.run(NetworkTimeout, args...)
	if err != nil {
		return fmt.Errorf("cannot fetch branches %v from %s: %w", branches, remote, err)
	}

	return nil
}

// GetDefaultBranch trả về default branch của remote, fallback về "main"
func (r Runner) GetDefaultBranch(remote string) (string, error) {
	// Thử phương pháp 1: git symbolic-ref refs/remotes/remote/HEAD
	out, err := r.run(NetworkTimeout, "symbolic-ref", fmt.Sprintf("refs/remotes/%s/HEAD", remote))
	if err == nil {
		// Output format: refs/remotes/origin/main
		parts := strings.Split(strings.TrimSpace(out), "/")
		if len(parts) >= 2 {
			return parts[len(parts)-1], nil
		}
	}

	// Thử phương pháp 2: git ls-remote --symref remote HEAD
	out, err = r.run(NetworkTimeout, "ls-remote", "--symref", remote, "HEAD")
	if err == nil {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ref: refs/heads/") {
				branch := strings.TrimPrefix(line, "ref: refs/heads/")
				if idx := strings.Index(branch, "\t"); idx > 0 {
					branch = branch[:idx]
				}
				return strings.TrimSpace(branch), nil
			}
		}
	}

	// Fallback cuối cùng: "main" nếu cả hai phương pháp thất bại
	return "main", fmt.Errorf("cannot determine default branch for %s, using fallback 'main': %w", remote, err)
}

// GetCurrentBranch trả về tên branch hiện tại, "HEAD" cho detached HEAD
func (r Runner) GetCurrentBranch() (string, error) {
	out, err := r.run(DefaultCmdTimeout, "branch", "--show-current")
	if err != nil {
		return "HEAD", fmt.Errorf("cannot get current branch, using fallback 'HEAD': %w", err)
	}

	branch := strings.TrimSpace(out)
	if branch == "" {
		// Detached HEAD state
		return "HEAD", nil
	}

	return branch, nil
}
