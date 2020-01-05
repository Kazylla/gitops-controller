package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/kazylla/gitops-controller/controllers/version"

	"github.com/go-logr/logr"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	plumbing_http "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type Config struct {
	ImagePath     string
	Repo          string
	Branch        string
	ReleaseBranch string
	Paths         []string
	CommitName    string
	CommitEmail   string
	Log           logr.Logger
	Username      string
	Password      string
}

type GitRepo struct {
	config   Config
	fs       billy.Filesystem
	repo     *git.Repository
	worktree *git.Worktree
	remote   *git.Remote
}

// NewGitRepo clones the specified git repository branch
func NewGitRepo(c Config) (*GitRepo, error) {
	gitRepo := &GitRepo{}

	// when deploying via PR, clone c.ReleaseBranch.
	// if there is no c.ReleaseBranch branch yet, clone c.Branch as c.ReleaseBranch
	branches := make([]string, 0)
	if c.Branch == "" {
		c.Branch = "master"
	}
	if c.ReleaseBranch != "" {
		branches = append(branches, c.ReleaseBranch)
	} else {
		c.ReleaseBranch = c.Branch
	}
	branches = append(branches, c.Branch)

	// clone git repo into inmem storage
	var err error
	gitRepo.fs = memfs.New()
	for i, b := range branches {
		cloneOptions := &git.CloneOptions{
			URL:           c.Repo,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", b)),
		}
		if c.Username != "" && c.Password != "" {
			cloneOptions.Auth = &plumbing_http.BasicAuth{
				Username: c.Username,
				Password: c.Password,
			}
		}
		gitRepo.repo, err = git.Clone(memory.NewStorage(), gitRepo.fs, cloneOptions)
		if err == plumbing.ErrReferenceNotFound && i < len(branches)-1 {
			continue
		}
		if err != nil {
			return nil, err
		}
		break
	}

	// checkout specific branch
	gitRepo.worktree, err = gitRepo.repo.Worktree()
	if err != nil {
		return nil, err
	}
	err = gitRepo.worktree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.ReferenceName(c.ReleaseBranch),
	})
	if err != nil {
		return nil, err
	}

	// get remote
	gitRepo.remote, err = gitRepo.repo.Remote("origin")
	if err != nil {
		return nil, err
	}

	gitRepo.config = c

	return gitRepo, nil
}

// readFile reads the entire file from the specified path
func (gitRepo *GitRepo) readFile(path string) ([]byte, error) {
	file, err := gitRepo.fs.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	readBuf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return readBuf, nil
}

// updateFile updates the file at the specified path
func (gitRepo *GitRepo) updateFile(path string, writeBuf []byte) error {
	file, err := gitRepo.fs.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write(writeBuf)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	_, err = gitRepo.worktree.Add(path)
	if err != nil {
		return err
	}
	return nil
}

// CommitTags creates a commit that updates specific image tags
func (gitRepo *GitRepo) CommitTags(imageVers []version.ImageVersion) (string, error) {
	log := gitRepo.config.Log.WithValues("git_repo", gitRepo.config.Repo)

	var latestTag string

	for _, v := range imageVers {
		registryTag := v.GetTag()
		latestTag = registryTag
		updated := false

		log.Info("processing", "tag", registryTag)

		for _, path := range gitRepo.config.Paths {

			log.Info("reading", "path", path)

			readBuf, err := gitRepo.readFile(path)
			if err != nil {
				return "", err
			}

			m := make(map[interface{}]interface{})
			err = yaml.Unmarshal(readBuf, &m)

			if _, ok := m["imageTags"]; !ok {
				log.Info("there is no imageTags, ignored", "path", path)
				continue
			}

			imageTags := m["imageTags"].([]interface{})
			for _, t := range imageTags {
				imageTag := t.(map[interface{}]interface{})
				imageName := imageTag["name"].(string)
				if imageName != gitRepo.config.ImagePath {
					// continue if image name is not target
					// (kustomize.yaml may have newTags for multiple images)
					log.V(1).Info("image name is not target", "target", gitRepo.config.ImagePath, "yaml", imageName)
					continue
				}

				// continue if the version on git repository is newer
				if _, ok := imageTag["newTag"]; ok {
					imageNewTag := imageTag["newTag"].(string)
					result, err := v.Compare(imageNewTag)
					if err != nil {
						return "", err
					}
					if result <= 0 {
						log.V(1).Info("this tag is older than current", "current", imageNewTag, "this tag", registryTag)
						continue
					}
				} else {
					log.Info("since newTag was not found, create it", "newTag", registryTag)
				}

				// update version
				imageTag["newTag"] = registryTag
				writeBuf, err := yaml.Marshal(m)
				if err != nil {
					return "", err
				}
				err = gitRepo.updateFile(path, writeBuf)
				if err != nil {
					return "", err
				}

				updated = true
				log.Info("updated", "path", path)
			}
		}

		if !updated {
			continue
		}

		commitLog := fmt.Sprintf("update imageTags to %s for %s by gitops-controller", registryTag, gitRepo.config.ImagePath)
		hash, prBranch, err := gitRepo.commitAndPush(registryTag, commitLog, gitRepo.config.CommitName, gitRepo.config.CommitEmail)
		if err != nil {
			return "", err
		}

		log.Info("new commit created", "tag", registryTag, "hash", hash)

		if prBranch != "" {
			pr := NewPR(gitRepo.config.Repo, gitRepo.config.Username, gitRepo.config.Password)
			if pr != nil {
				err = pr.CreatePR(registryTag, prBranch)
				if err != nil {
					return "", err
				}
				log.Info("new PR created", "tag", registryTag, "hash", hash)
			}
		}
	}

	return latestTag, nil
}

// commitAndPush creates one commit from the work tree and pushes to remote origin
func (gitRepo *GitRepo) commitAndPush(tag, commitLog, name, email string) (string, string, error) {
	commit, err := gitRepo.worktree.Commit(commitLog, &git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", "", err
	}
	err = gitRepo.repo.Storer.SetReference(plumbing.NewReferenceFromStrings(gitRepo.config.ReleaseBranch, commit.String()))
	if err != nil {
		return "", "", err
	}

	var prBranch string
	branches := []string{gitRepo.config.ReleaseBranch}
	if gitRepo.config.Branch != gitRepo.config.ReleaseBranch {
		prBranch = fmt.Sprintf("%s-%s", gitRepo.config.ReleaseBranch, tag)
		branches = append(branches, prBranch)
	}

	for _, b := range branches {
		pushOptions := &git.PushOptions{
			Progress: ioutil.Discard,
			RefSpecs: []config.RefSpec{
				config.RefSpec(plumbing.ReferenceName(fmt.Sprintf("%s:refs/heads/%s", gitRepo.config.ReleaseBranch, b))),
			},
		}
		if gitRepo.config.Username != "" && gitRepo.config.Password != "" {
			pushOptions.Auth = &plumbing_http.BasicAuth{
				Username: gitRepo.config.Username,
				Password: gitRepo.config.Password,
			}
		}
		err = gitRepo.remote.Push(pushOptions)
		if err != nil {
			return "", "", err
		}
	}

	return commit.String(), prBranch, nil
}
