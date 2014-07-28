package main

import (
	"errors"
	"os"
	"strings"
)

type RepositoryStore interface {
	GetRepository(string) (Repository, error)
}

type GitRepositoryStore struct {
	path string
}

var _ RepositoryStore = &GitRepositoryStore{}

func (s *GitRepositoryStore) GetRepository(path string) (Repository, error) {
	path = strings.Trim(path, "/")
	fullPath := s.fullRepoPath(path)

	if !s.isGitRepo(fullPath) {
		path = path + "/.git"
		fullPath = s.fullRepoPath(path)

		if !s.isGitRepo(fullPath) {
			return nil, errors.New(path + " is not a path to a git repository")
		}
	}

	return &GitRepository{fullPath}, nil
}

func (s *GitRepositoryStore) fullRepoPath(path string) string {
	return s.path + "/" + path
}

func (s *GitRepositoryStore) isGitRepo(path string) bool {
	refsPath := path + "/refs"

	_, err := os.Stat(refsPath)

	if err == nil {
		return true
	}

	return false
}
