package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func main() {
	defaultBasePath := getDefaultBasePath()
	basePath := flag.String("path", defaultBasePath, "Base path to search for git repositories")
	flag.Parse()
	resolvedPath, err := expandPath(*basePath)
	if err != nil {
		fmt.Printf("%sError resolving path %s: %v%s\n", colorRed, *basePath, err, colorReset)
		return
	}
	resolvedPath = sanitizePath(resolvedPath)
	repoCount := 0
	attentionRepos := []string{}
	fmt.Printf("%sSearching for git repositories in %s...%s\n", colorCyan, resolvedPath, colorReset)
	err = filepath.Walk(resolvedPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && isGitRepository(path) {
			repoCount++
			if checkRepoStatus(path) {
				attentionRepos = append(attentionRepos, path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("%sError walking the path: %v%s\n", colorRed, err, colorReset)
		return
	}
	if repoCount == 0 {
		fmt.Printf("%sNo git repositories found.%s\n", colorYellow, colorReset)
	} else if len(attentionRepos) == 0 {
		fmt.Printf("%sAll %d repositories are clean%s\n", colorGreen, repoCount, colorReset)
	} else {
		fmt.Printf("%sThe following repositories have uncommitted changes:%s\n", colorYellow, colorReset)
		for _, repo := range attentionRepos {
			fmt.Printf("%s - %s%s\n", colorYellow, repo, colorReset)
		}
	}
}

func getDefaultBasePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(string(colorRed), "Error getting user home directory: %v\n", err, colorReset)
		os.Exit(1)
	}
	return homeDir
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[1:]), nil
	}
	return filepath.Abs(path)
}

func sanitizePath(path string) string {
	path = strings.Trim(path, `"'`)
	return strings.TrimRight(path, `\/`)
}

func isGitRepository(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return !os.IsNotExist(err)
}

func checkRepoStatus(path string) bool {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%sError checking repository status %s: %v%s\n", colorRed, path, err, colorReset)
		return false
	}
	return strings.TrimSpace(out.String()) != ""
}
