package main

import (
	"os"
	"os/exec"
	"testing"
)

func execCommand(command string) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func errToBool(err error) bool {
	if err == nil {
		return false
	}

	return true
}

func TestGitRepositoryStore_GetRepository(t *testing.T) {
	execCommand("cd fixtures/depot/project-x && pwd && ls -la && rm -rf non-bare-repo-a && git clone repo-a.git non-bare-repo-a")

	var tests = []struct {
		repoPath       string
		actualRepoPath string
		error          bool
	}{
		{"project-x/repo-a.git", "fixtures/depot/project-x/repo-a.git", false},
		{"project-x/non-bare-repo-a", "fixtures/depot/project-x/non-bare-repo-a/.git", false},
		{"/project-y/repo-b.git/", "fixtures/depot/project-y/repo-b.git", false},
		{"project-x/non-existent.git", "", true},
		{"project-z/repo-b.git", "", true},
		{"", "", true},
	}

	store := &GitRepositoryStore{"fixtures/depot"}

	for _, test := range tests {
		repo, err := store.GetRepository(test.repoPath)

		if errToBool(err) != test.error {
			t.Errorf("expected error to be %v for %v", test.error, test)
		}

		if err == nil && repo.Path() != test.actualRepoPath {
			t.Errorf("expected actualRepoPath to be %v for %v, got %v", test.actualRepoPath, test, repo.Path())
		}
	}
}
