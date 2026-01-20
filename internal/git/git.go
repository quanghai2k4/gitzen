package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrGitNotFound     = errors.New("git not found")
	ErrNotARepository  = errors.New("not a git repository")
	DefaultCmdTimeout  = 3 * time.Second
	DefaultDiffTimeout = 10 * time.Second
)

type Runner struct {
	RepoRoot string
}

func LookPath() error {
	if _, err := exec.LookPath("git"); err != nil {
		return ErrGitNotFound
	}
	return nil
}

func DetectRepoRoot(repoPath string) (string, error) {
	args := []string{"rev-parse", "--show-toplevel"}
	out, err := runRaw(repoPath, DefaultCmdTimeout, args...)
	if err != nil {
		return "", ErrNotARepository
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return "", ErrNotARepository
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root, nil
	}
	return abs, nil
}

func New(repoRoot string) Runner {
	return Runner{RepoRoot: repoRoot}
}

func (r Runner) StatusPorcelainZ() ([]byte, error) {
	return r.runBytes(DefaultCmdTimeout, "status", "--porcelain=v1", "-z")
}

func (r Runner) LogOneline() (string, error) {
	return r.run(DefaultCmdTimeout, "log", "--oneline", "--decorate", "-n", "200")
}

// Reflog returns the reflog entries
func (r Runner) Reflog() (string, error) {
	return r.run(DefaultCmdTimeout, "reflog", "-n", "100")
}

func (r Runner) DiffFile(path string, staged bool) (string, error) {
	if staged {
		return r.run(DefaultDiffTimeout, "diff", "--staged", "--", path)
	}
	return r.run(DefaultDiffTimeout, "diff", "--", path)
}

func (r Runner) ShowCommit(hash string) (string, error) {
	return r.run(DefaultDiffTimeout, "show", hash)
}

func (r Runner) Add(path string) error {
	_, err := r.run(DefaultCmdTimeout, "add", "--", path)
	return err
}

func (r Runner) RestoreStaged(path string) error {
	_, err := r.run(DefaultCmdTimeout, "restore", "--staged", "--", path)
	return err
}

func (r Runner) Commit(message string) (string, error) {
	return r.run(DefaultCmdTimeout, "commit", "-m", message)
}

func (r Runner) CurrentBranch() (string, error) {
	return r.run(DefaultCmdTimeout, "rev-parse", "--abbrev-ref", "HEAD")
}

// Branch represents a git branch
type Branch struct {
	Name      string
	IsCurrent bool
	IsRemote  bool
}

// ListBranches returns all local branches
func (r Runner) ListBranches() ([]Branch, error) {
	out, err := r.run(DefaultCmdTimeout, "branch", "--format=%(HEAD)%(refname:short)")
	if err != nil {
		return nil, err
	}
	var branches []Branch
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		isCurrent := false
		name := line
		if len(line) > 0 && line[0] == '*' {
			isCurrent = true
			name = line[1:]
		}
		branches = append(branches, Branch{
			Name:      strings.TrimSpace(name),
			IsCurrent: isCurrent,
			IsRemote:  false,
		})
	}
	return branches, nil
}

// StashEntry represents a git stash entry
type StashEntry struct {
	Index   int
	Ref     string // stash@{0}
	Message string
}

// ListStash returns all stash entries
func (r Runner) ListStash() ([]StashEntry, error) {
	out, err := r.run(DefaultCmdTimeout, "stash", "list")
	if err != nil {
		// stash list returns error if no stash, but also might just be empty
		return nil, nil
	}
	var entries []StashEntry
	for i, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		// Format: stash@{0}: WIP on branch: message
		// or: stash@{0}: On branch: message
		parts := strings.SplitN(line, ": ", 2)
		ref := parts[0]
		msg := ""
		if len(parts) > 1 {
			msg = parts[1]
		}
		entries = append(entries, StashEntry{
			Index:   i,
			Ref:     ref,
			Message: msg,
		})
	}
	return entries, nil
}

// ShowStash shows the diff for a stash entry
func (r Runner) ShowStash(ref string) (string, error) {
	return r.run(DefaultDiffTimeout, "stash", "show", "-p", ref)
}

// DiffBranch shows the diff between current HEAD and a branch
func (r Runner) DiffBranch(branch string) (string, error) {
	return r.run(DefaultDiffTimeout, "diff", branch+"...HEAD")
}

// ========== HIGH PRIORITY FEATURES ==========

// DiscardFile discards changes in a file (unstaged changes only)
func (r Runner) DiscardFile(path string) error {
	_, err := r.run(DefaultCmdTimeout, "checkout", "--", path)
	return err
}

// DiscardUntracked removes an untracked file
func (r Runner) DiscardUntracked(path string) error {
	_, err := r.run(DefaultCmdTimeout, "clean", "-f", "--", path)
	return err
}

// Pull fetches and merges from remote
func (r Runner) Pull() (string, error) {
	return r.run(30*time.Second, "pull")
}

// Push pushes to remote
func (r Runner) Push() (string, error) {
	return r.run(30*time.Second, "push")
}

// PushSetUpstream pushes and sets upstream for new branch
func (r Runner) PushSetUpstream(remote, branch string) (string, error) {
	return r.run(30*time.Second, "push", "-u", remote, branch)
}

// CheckoutBranch switches to a branch
func (r Runner) CheckoutBranch(branch string) (string, error) {
	return r.run(DefaultCmdTimeout, "checkout", branch)
}

// CreateBranch creates a new branch and switches to it
func (r Runner) CreateBranch(name string) (string, error) {
	return r.run(DefaultCmdTimeout, "checkout", "-b", name)
}

// ========== MEDIUM PRIORITY FEATURES ==========

// Fetch fetches from remote
func (r Runner) Fetch() (string, error) {
	return r.run(30*time.Second, "fetch", "--all", "--prune")
}

// DeleteBranch deletes a local branch
func (r Runner) DeleteBranch(name string) error {
	_, err := r.run(DefaultCmdTimeout, "branch", "-d", name)
	return err
}

// DeleteBranchForce force deletes a local branch
func (r Runner) DeleteBranchForce(name string) error {
	_, err := r.run(DefaultCmdTimeout, "branch", "-D", name)
	return err
}

// StashApply applies a stash entry without removing it
func (r Runner) StashApply(ref string) error {
	_, err := r.run(DefaultCmdTimeout, "stash", "apply", ref)
	return err
}

// StashPop applies and removes a stash entry
func (r Runner) StashPop(ref string) error {
	_, err := r.run(DefaultCmdTimeout, "stash", "pop", ref)
	return err
}

// StashDrop removes a stash entry
func (r Runner) StashDrop(ref string) error {
	_, err := r.run(DefaultCmdTimeout, "stash", "drop", ref)
	return err
}

// CommitAmend amends the last commit with staged changes
func (r Runner) CommitAmend(message string) (string, error) {
	if message == "" {
		return r.run(DefaultCmdTimeout, "commit", "--amend", "--no-edit")
	}
	return r.run(DefaultCmdTimeout, "commit", "--amend", "-m", message)
}

// ResetSoftHead resets to previous commit, keeping changes staged
func (r Runner) ResetSoftHead(n int) error {
	_, err := r.run(DefaultCmdTimeout, "reset", "--soft", fmt.Sprintf("HEAD~%d", n))
	return err
}

// ResetMixedHead resets to previous commit, keeping changes unstaged
func (r Runner) ResetMixedHead(n int) error {
	_, err := r.run(DefaultCmdTimeout, "reset", "--mixed", fmt.Sprintf("HEAD~%d", n))
	return err
}

// GetRemote returns the default remote name (usually "origin")
func (r Runner) GetRemote() (string, error) {
	out, err := r.run(DefaultCmdTimeout, "remote")
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", errors.New("no remote configured")
	}
	// Prefer "origin" if exists
	for _, line := range lines {
		if line == "origin" {
			return "origin", nil
		}
	}
	return lines[0], nil
}

// HasUpstream checks if current branch has upstream configured
func (r Runner) HasUpstream() bool {
	_, err := r.run(DefaultCmdTimeout, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	return err == nil
}

func (r Runner) run(timeout time.Duration, args ...string) (string, error) {
	out, err := runRaw(r.RepoRoot, timeout, args...)
	if err != nil {
		return "", err
	}
	return out, nil
}

func (r Runner) runBytes(timeout time.Duration, args ...string) ([]byte, error) {
	out, err := runRawBytes(r.RepoRoot, timeout, args...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func runRaw(repoRoot string, timeout time.Duration, args ...string) (string, error) {
	b, err := runRawBytes(repoRoot, timeout, args...)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func runRawBytes(repoRoot string, timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	if repoRoot != "" {
		cmd.Dir = repoRoot
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("git %s: timeout", strings.Join(args, " "))
	}
	if err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = strings.TrimSpace(stdout.String())
		}
		if errText == "" {
			errText = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), errText)
	}

	return stdout.Bytes(), nil
}
