package git

import (
	"fmt"
	"strings"
)

type PR interface {
	CreatePR(tag, branch string) error
}

func NewPR(repo, username, password string) PR {
	repoURL, err := parseRepoURL(repo)
	if err != nil {
		return nil
	}
	switch repoURL.Host {
	case "github.com":
		return NewGithubPR(repo, username, password)
	default:
		return nil
	}
}

type RepoURL struct {
	Host     string
	Owner    string
	RepoName string
}

// parseRepoURL parses GitHub repository repoURL string for https or git protocol
func parseRepoURL(repo string) (*RepoURL, error) {
	repoURL := &RepoURL{}

	switch {
	case strings.HasPrefix(repo, "https://"):
		parts := strings.Split(repo, "/")
		if len(parts) == 5 {
			repoURL.Host = parts[2]
			repoURL.Owner = parts[3]
			repoURL.RepoName = parts[4]
		}
	case strings.HasPrefix(repo, "git@"):
		parts := strings.Split(repo, ":")
		if len(parts) == 2 {
			parts1 := strings.Split(parts[0], "@")
			if len(parts1) == 2 {
				repoURL.Host = parts1[1]
			}
			parts2 := strings.Split(parts[1], "/")
			if len(parts2) == 2 {
				repoURL.Owner = parts2[0]
				repoURL.RepoName = parts2[1]
			}
		}
	}

	if repoURL.Host == "" || repoURL.Owner == "" || repoURL.RepoName == "" || !strings.HasSuffix(repoURL.RepoName, ".git") {
		return nil, fmt.Errorf("invalid repoURL string")
	}

	repoURL.RepoName = strings.TrimRight(repoURL.RepoName, ".git")

	return repoURL, nil
}
