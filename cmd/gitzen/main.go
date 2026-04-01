package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitzen/internal/app"
	"gitzen/internal/updater"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse arguments manually to support both short and long flags
	args := os.Args[1:]

	var repoPath string
	var showVersion, showHelp, uninstallFlag, updateFlag, updateDryRun bool

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			showHelp = true
		case arg == "-v" || arg == "--version":
			showVersion = true
		case arg == "-u" || arg == "--update":
			updateFlag = true
		case arg == "--update-dry-run":
			updateFlag = true
			updateDryRun = true
		case arg == "--uninstall":
			uninstallFlag = true
		case arg == "-r" || arg == "--repo":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				repoPath = args[i+1]
				i++ // Skip next argument as it's the value
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a value\n", arg)
				printUsage()
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--repo="):
			repoPath = strings.TrimPrefix(arg, "--repo=")
		case strings.HasPrefix(arg, "-r="):
			repoPath = strings.TrimPrefix(arg, "-r=")
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "Error: unknown flag %s\n", arg)
			printUsage()
			os.Exit(1)
		default:
			// Treat as positional argument for repo path
			if repoPath == "" {
				repoPath = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: multiple repository paths specified\n")
				printUsage()
				os.Exit(1)
			}
		}
	}

	if showHelp {
		printUsage()
		return
	}

	if showVersion {
		printVersion()
		return
	}

	if uninstallFlag {
		uninstall()
		return
	}

	if updateFlag {
		runUpdate(updateDryRun)
		return
	}

	exitCode := app.Run(app.Options{
		RepoPath: repoPath,
		Version:  version,
		Commit:   commit,
	})
	os.Exit(exitCode)
}

// printVersion displays version information
func printVersion() {
	fmt.Printf("GitZen %s\n", version)
	if commit != "none" && commit != "" {
		fmt.Printf("commit: %s\n", commit)
	}
	if date != "unknown" && date != "" {
		fmt.Printf("built:  %s\n", date)
	}
	fmt.Printf("\nA TUI Git client inspired by lazygit\n")
	fmt.Printf("Repository: https://github.com/quanghai2k4/gitzen\n")
}

// printUsage displays help information
func printUsage() {
	fmt.Printf(`GitZen - A TUI Git client inspired by lazygit

Usage:
  gitzen [flags] [repository-path]

Flags:
  -h, --help          Show this help message and exit
  -v, --version       Show version information and exit
  -u, --update        Check for and install updates from GitHub
      --update-dry-run Check for updates without installing (dry run)
  -r, --repo <path>   Specify git repository path
      --uninstall     Uninstall GitZen from system

Arguments:
  repository-path     Path to git repository (optional, defaults to current directory)

Examples:
  gitzen                      # Run in current directory
  gitzen /path/to/repo        # Run in specific repository  
  gitzen -r ~/projects/app    # Run with --repo flag
  gitzen --repo=/tmp/project  # Run with --repo= format
  gitzen -v                   # Show version (short flag)
  gitzen --version            # Show version (long flag)
  gitzen -u                   # Check for updates (short flag)
  gitzen --update             # Check for updates (long flag)
  gitzen --update-dry-run     # Check what would be updated (dry run)
  gitzen -h                   # Show this help (short flag)
  gitzen --help               # Show this help (long flag)
  gitzen --uninstall          # Uninstall GitZen

For more information and documentation:
  Repository: https://github.com/quanghai2k4/gitzen
  Issues:     https://github.com/quanghai2k4/gitzen/issues
`)
}

// uninstall removes the gitzen binary from the system
func uninstall() {
	fmt.Println("GitZen Uninstaller")
	fmt.Println("==================")

	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not find executable path: %v\n", err)
		os.Exit(1)
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	fmt.Printf("Removing GitZen from: %s\n", realPath)
	fmt.Print("Are you sure? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Uninstall cancelled.")
		os.Exit(0)
	}

	// Check if we have permission to delete
	if err := os.Remove(realPath); err != nil {
		if os.IsPermission(err) {
			fmt.Fprintf(os.Stderr, "Permission denied. Try: sudo gitzen --uninstall\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: Could not remove %s: %v\n", realPath, err)
		os.Exit(1)
	}

	fmt.Println("✅ GitZen has been uninstalled successfully!")
	fmt.Println("Thank you for using GitZen!")
}

// runUpdate handles the update process
func runUpdate(dryRun bool) {
	fmt.Println("GitZen Updater")
	fmt.Println("==============")

	u := updater.NewUpdater(version)

	// Check for updates first
	fmt.Println("Checking for updates...")
	latest, err := u.CheckForUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for updates: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Current version: v%s\n", version)

	if latest == nil {
		fmt.Printf("✓ Already up to date! (v%s)\n", version)
		return
	}

	fmt.Printf("Latest version:  v%s\n", latest.TagName)
	fmt.Printf("Release date:    %s\n", latest.PublishedAt)

	// Show release notes if available
	if latest.Body != "" {
		fmt.Printf("\nRelease Notes:\n%s\n\n", strings.TrimSpace(latest.Body))
	}

	if dryRun {
		fmt.Printf("🔍 DRY RUN: Would update from v%s to v%s\n", version, latest.TagName)
		fmt.Println("Use --update to perform the actual update.")
		return
	}

	// Confirm update
	fmt.Print("Do you want to update? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Update cancelled.")
		return
	}

	// Perform update
	options := updater.UpdateOptions{
		DryRun:  false,
		Force:   false,
		Backup:  true,
		Verbose: true,
	}

	result, err := u.Update(options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		
		// Show backup info if available
		if result.BackupPath != "" {
			fmt.Fprintf(os.Stderr, "Your previous version was backed up to: %s\n", result.BackupPath)
		}
		os.Exit(1)
	}

	fmt.Println(result.Message)
	if result.BackupPath != "" {
		fmt.Printf("Previous version backed up to: %s\n", result.BackupPath)
	}
}
