package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func extractArchive(archivePath, format, destDirPath string) {
	var extractCommand string

	if format == "zip" {
		extractCommand = "unzip " + archivePath
	} else if format == "tar.gz" {
		extractCommand = "tar xzf " + archivePath
	}

	command := "cd " + destDirPath + " && " + extractCommand

	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()

	if err != nil {
		fmt.Println(output)
	}
}

func TestGitRepository_ResolveRef(t *testing.T) {
	var tests = []struct {
		repoPath string
		ref      string
		commit   string
		error    bool
	}{
		{"fixtures/depot/project-y/repo-b.git", "next", "65c9c42f640cd944972ad9465bba09f5c6a6bfe4", false},
		{"fixtures/depot/project-y/repo-b.git", "65c9c42f640cd944972ad9465bba09f5c6a6bfe4", "65c9c42f640cd944972ad9465bba09f5c6a6bfe4", false},
		{"fixtures/depot/project-y/repo-b.git", "nonexistent", "", true},
		{"fixtures/depot/non-existent/nope.git", "nonexistent", "", true},
	}

	for _, test := range tests {
		repo := &GitRepository{test.repoPath}

		commit, err := repo.ResolveRef(test.ref)

		if test.error && err == nil {
			t.Errorf("expected error for %v", test.ref)
		}

		if !test.error && err != nil {
			t.Errorf("didn't expect error for %v", test.ref)
		}

		if test.commit != commit {
			t.Errorf("expected %v for %v, got %v", test.commit, test.ref, commit)
		}
	}
}

func TestGitRepository_Archive(t *testing.T) {
	var tests = []struct {
		repoPath        string
		commit          string
		prefix          string
		format          string
		archiveFilename string
		expectedFile    string
		expectedContent string
		error           bool
	}{
		{
			"fixtures/depot/project-y/repo-b.git",
			"master", "prefix-master/", "zip", "a.zip", "INSTALL", "", false,
		},
		{
			"fixtures/depot/project-y/repo-b.git",
			"next", "prefix-next/", "tar.gz", "b.tar.gz", "INSTALL", "fixed\n", false,
		},
		{
			"fixtures/depot/project-y/repo-b.git",
			"nonexistent", "prefix/", "tar.gz", "c.tar.gz", "", "", true,
		},
		{
			"fixtures/depot/project-y/repo-b.git",
			"next", "prefix/", "rar", "d.rar", "", "", true,
		},
		{
			"fixtures/depot/non-existent/nope.git",
			"master", "prefix/", "zip", "e.zip", "", "", true,
		},
	}

	tmpDir, _ := ioutil.TempDir("", "")

	for _, test := range tests {
		repo := &GitRepository{test.repoPath}

		outPath := tmpDir + "/" + test.archiveFilename
		err := repo.Archive(test.commit, test.prefix, test.format, outPath)

		if test.error && err == nil {
			t.Errorf("expected error for %v", test)
		}

		if !test.error && err != nil {
			t.Errorf("didn't expect error for %v", test)
		}

		if err == nil {
			extractArchive(outPath, test.format, tmpDir)
			bytes, err := ioutil.ReadFile(tmpDir + "/" + test.prefix + "/" + test.expectedFile)

			if err == nil {
				content := string(bytes)

				if content != test.expectedContent {
					t.Errorf("expected file %v to have content \"%v\" for %v, got \"%v\"", test.expectedFile, test.expectedContent, test, content)
				}
			} else {
				t.Errorf("expected file %v to be there for %v", test.expectedFile, test)
			}
		}
	}

	os.RemoveAll(tmpDir)
}
