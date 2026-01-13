package main

import (
	"flag"
	"fmt"
	"os"

	"gitzen/internal/app"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	repoFlag := flag.String("repo", "", "path to git repository")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gitzen %s (%s)\n", version, commit)
		return
	}

	exitCode := app.Run(app.Options{
		RepoPath: *repoFlag,
		Version:  version,
		Commit:   commit,
	})
	os.Exit(exitCode)
}
