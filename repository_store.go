package main

import (
	"errors"
	"os"
	"strings"
)

type RepositoryStore interface {
	GetRepository(path string) (Repository, error)
}

type GitRepositoryStore struct {
	path string
}

func (d *GitRepositoryStore) GetRepository(path string) (Repository, error) {
	path = strings.Trim(path, "/")
	fullPath := d.fullRepoPath(path)

	if !d.isGitRepo(fullPath) {
		path = path + "/.git"
		fullPath = d.fullRepoPath(path)

		if !d.isGitRepo(fullPath) {
			return nil, errors.New(path + " is not a path to a git repository")
		}
	}

	return &GitRepository{fullPath}, nil
}

func (d *GitRepositoryStore) fullRepoPath(path string) string {
	return d.path + "/" + path
}

func (d *GitRepositoryStore) isGitRepo(path string) bool {
	refsPath := path + "/refs"

	_, err := os.Stat(refsPath)

	if err == nil {
		return true
	}

	return false
}
