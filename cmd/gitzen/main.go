package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gitzen/internal/app"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	repoFlag := flag.String("repo", "", "path to git repository")
	showVersion := flag.Bool("version", false, "print version")
	uninstallFlag := flag.Bool("uninstall", false, "uninstall gitzen")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gitzen %s (%s)\n", version, commit)
		return
	}

	if *uninstallFlag {
		uninstall()
		return
	}

	exitCode := app.Run(app.Options{
		RepoPath: *repoFlag,
		Version:  version,
		Commit:   commit,
	})
	os.Exit(exitCode)
}

// uninstall removes the gitzen binary from the system
func uninstall() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not find executable path: %v\n", err)
		os.Exit(1)
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	fmt.Printf("Uninstalling gitzen from %s\n", realPath)

	// Check if we have permission to delete
	if err := os.Remove(realPath); err != nil {
		if os.IsPermission(err) {
			fmt.Fprintf(os.Stderr, "Permission denied. Try: sudo gitzen --uninstall\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: could not remove %s: %v\n", realPath, err)
		os.Exit(1)
	}

	fmt.Println("GitZen has been uninstalled successfully!")
}
