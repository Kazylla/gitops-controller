package git

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
)

type GithubPR struct {
	RepoURL  string
	Username string
	Password string
}

func NewGithubPR(repoURL, username, password string) *GithubPR {
	return &GithubPR{
		RepoURL:  repoURL,
		Username: username,
		Password: password,
	}
}

// CreatePR creates a pull request for the specified branch
func (pr *GithubPR) CreatePR(tag, branch string) error {
	tp := github.BasicAuthTransport{
		Username: pr.Username,
		Password: pr.Password,
	}
	client := github.NewClient(tp.Client())

	newPR := &github.NewPullRequest{
		Title:               github.String(fmt.Sprintf("Release Candidate: %s", tag)),
		Head:                github.String(branch),
		Base:                github.String("master"),
		Body:                github.String(fmt.Sprintf("If you want to deploy version %s, please merge this PR", tag)),
		MaintainerCanModify: github.Bool(true),
	}

	repoURL, err := parseRepoURL(pr.RepoURL)
	if err != nil {
		return err
	}

	_, _, err = client.PullRequests.Create(context.Background(), repoURL.Owner, repoURL.RepoName, newPR)
	if err != nil {
		return err
	}

	return nil
}
