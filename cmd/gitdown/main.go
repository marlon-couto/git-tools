package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Repository struct {
	Name string `json:"name"`
	Url  string `json:"clone_url"`
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func main() {
	username := flag.String("username", "", "GitHub username")
	path := flag.String("path", ".", "Path to clone repositories")
	flag.Parse()

	if *username == "" {
		fmt.Printf("%sEnter your GitHub username: %s", colorCyan, colorReset)
		var inputUsername string
		fmt.Scanln(&inputUsername)
		*username = inputUsername
	}

	resolvedPath, err := expandPath(*path)
	if err != nil {
		fmt.Printf("%sError resolving path %s: %v%s\n", colorRed, *path, err, colorReset)
		return
	}
	resolvedPath = sanitizePath(resolvedPath)
	if err := os.MkdirAll(resolvedPath, os.ModePerm); err != nil {
		fmt.Printf("%sError creating directory %s: %v%s\n", colorRed, resolvedPath, err, colorReset)
		return
	}

	repos, err := getRepositories(*username)
	if err != nil {
		fmt.Printf("%sError getting repositories: %v%s\n", colorRed, err, colorReset)
		return
	}

	for _, repo := range repos {
		clonePath := filepath.Join(resolvedPath, repo.Name)
		if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
			fmt.Printf("%sRepository %s already exists, skipping...%s\n", colorYellow, repo.Name, colorReset)
			continue
		}

		fmt.Printf("%sCloning %s into %s...%s\n", colorCyan, repo.Name, clonePath, colorReset)
		if err := cloneRepository(repo.Url, repo.Name, resolvedPath); err != nil {
			fmt.Printf("%sError cloning repository %s: %v%s\n", colorRed, repo.Name, err, colorReset)
		}
	}
	fmt.Printf("%sAll repositories cloned successfully!%s\n", colorGreen, colorReset)
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

func getRepositories(username string) ([]Repository, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var repos []Repository
	if err := json.NewDecoder(res.Body).Decode(&repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func cloneRepository(url, name, path string) error {
	clonePath := filepath.Join(path, name)
	cmd := exec.Command("git", "clone", url, clonePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
