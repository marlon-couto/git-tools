package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	basePath := flag.String("path", ".", "Base path to search for git repositories")
	flag.Parse()

	repos, err := findGitRepos(*basePath)
	if err != nil {
		fmt.Printf("Error scanning repositories: %v\n", err)
		return
	}

	if len(repos) == 0 {
		fmt.Println("No Git repository found.")
	} else {
		fmt.Println("Git repositories found:")
		for _, repo := range repos {
			fmt.Println(repo)
		}
	}
}

func findGitRepos(basePath string) ([]string, error) {
	var repos []string

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Permission denied for %s. Skipping...\n", path)
			return nil
		}

		if info.IsDir() && filepath.Base(path) == ".git" {
			repos = append(repos, filepath.Dir(path))
		}
		return nil
	})

	return repos, err
}
