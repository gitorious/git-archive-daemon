package main

import (
	"errors"
	"fmt"
	"testing"
)

type testRepository struct {
	commit string
}

func (r *testRepository) Path() string {
	return "/the/path.git"
}

func (r *testRepository) ResolveRef(ref string) (string, error) {
	if r.commit != "" {
		return r.commit, nil
	}

	return "", errors.New("no ref")
}

func (r *testRepository) Archive(commit, prefix, format, outPath string) error {
	return nil
}

type testRepositoryStore struct {
	repository Repository
}

func (r *testRepositoryStore) GetRepository(path string) (Repository, error) {
	if r.repository != nil {
		return r.repository, nil
	}

	return nil, errors.New("no repo")
}

func handleRequest(request *ArchiveRequest, archiveError error) {
	if archiveError != nil {
		request.ResultChan <- &ArchiveResult{"", archiveError}
	} else {
		job := request.Job
		path := fmt.Sprintf(
			"%v/%v-%v-%v",
			job.RepoPath,
			job.Commit,
			job.Prefix,
			job.Filename,
		)
		request.ResultChan <- &ArchiveResult{path, nil}
	}
}

func TestArchiveGenerator_GenerateArchive(t *testing.T) {
	requestQueue := make(chan *ArchiveRequest)
	archiveError := errors.New("damn")

	var tests = []struct {
		repository    Repository
		archiveError  error
		expectedPath  string
		expectedError error
	}{
		{nil, nil, "", REPOSITORY_NOT_FOUND},
		{&testRepository{}, nil, "", REF_NOT_FOUND},
		{&testRepository{commit: "baadf00d"}, nil, "/the/path.git/baadf00d-prefix-6f7fc8032e76ae69e4b3c673ce57da527c84d4e1.zip", nil},
		{&testRepository{commit: "baadf00d"}, archiveError, "", archiveError},
	}

	for _, test := range tests {
		go func() {
			handleRequest(<-requestQueue, test.archiveError)
		}()

		generator := ArchiveGenerator{
			&testRepositoryStore{test.repository},
			requestQueue,
		}

		path, err := generator.GenerateArchive("repo/path", "deadbeef", "prefix", "zip")

		if path != test.expectedPath {
			t.Errorf(`expected path "%v", got "%v"`, test.expectedPath, path)
		}

		if err != test.expectedError {
			t.Errorf("expected error %v, got %v", test.expectedError, err)
		}
	}
}

func TestNewArchiveJob(t *testing.T) {
	var tests = []struct {
		repoPath         string
		commit           string
		prefix           string
		format           string
		expectedFilename string // (sha1(repoPath + commit + prefix)) + .format
	}{
		{"ab", "cafebabe", "prefix", "zip", "1919e806fd1a8351c4c7ff7011951af8ffed3506.zip"},
		{"ab.git", "cafebabe", "prefix/", "zip", "c8c0e5b45c70bf627f332b6488c651859c7d390d.zip"},
		{"ab-cd/ef", "cafebabe", "the/prefix/", "tar.gz", "cc19a57c7ffe13da071f354e87b6639c1bbe043f.tar.gz"},
	}

	for _, test := range tests {
		job := NewArchiveJob(test.repoPath, test.commit, test.prefix, test.format)

		if job.Filename != test.expectedFilename {
			t.Errorf(`expected "%v", got "%v"`, test.expectedFilename, job.Filename)
		}
	}
}
