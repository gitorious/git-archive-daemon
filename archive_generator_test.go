package main

import (
	"errors"
	"fmt"
	"testing"
)

type testRepository struct {
	commit       string
	archiveError error
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
	return r.archiveError
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

func handleRequest(request Request, archiveError error) {
	archiveRequest := request.(*ArchiveRequest)

	if archiveError != nil {
		request.ResultChan() <- &ArchiveResult{"", archiveError}
	} else {
		path := fmt.Sprintf(
			"%v/%v-%v-%v.%v",
			archiveRequest.archiveDir,
			archiveRequest.relativeRepoPath,
			archiveRequest.commit,
			archiveRequest.prefix,
			archiveRequest.format,
		)
		request.ResultChan() <- &ArchiveResult{path, nil}
	}
}

func TestArchiveGenerator_GenerateArchive(t *testing.T) {
	requestQueue := make(chan Request)
	archiveError := errors.New("damn")

	var tests = []struct {
		repository    Repository
		archiveError  error
		expectedPath  string
		expectedError error
	}{
		{nil, nil, "", REPOSITORY_NOT_FOUND},
		{&testRepository{}, nil, "", REF_NOT_FOUND},
		{&testRepository{commit: "baadf00d"}, nil, "/tmp/dir/repo/path-baadf00d-prefix.zip", nil},
		{&testRepository{commit: "baadf00d"}, archiveError, "", archiveError},
	}

	for _, test := range tests {
		go func() {
			handleRequest(<-requestQueue, test.archiveError)
		}()

		generator := ArchiveGenerator{
			&testRepositoryStore{test.repository},
			"/tmp/dir",
			requestQueue,
		}

		path, err := generator.GenerateArchive("repo/path", "deadbeef", "prefix", "zip")

		if path != test.expectedPath {
			t.Errorf("expected path \"%v\", got \"%v\"", test.expectedPath, path)
		}

		if err != test.expectedError {
			t.Errorf("expected error %v, got %v", test.expectedError, err)
		}
	}
}

func TestArchiveRequest_Hash(t *testing.T) {
	var tests = []struct {
		relativeRepoPath string
		commit           string
		prefix           string
		format           string
		expectedHash     string // (sha1(relativeRepoPath + commit + prefix)) + .format
	}{
		{"ab", "cafebabe", "prefix", "zip", "1919e806fd1a8351c4c7ff7011951af8ffed3506.zip"},
		{"ab.git", "cafebabe", "prefix/", "zip", "c8c0e5b45c70bf627f332b6488c651859c7d390d.zip"},
		{"ab-cd/ef", "cafebabe", "the/prefix/", "tar.gz", "cc19a57c7ffe13da071f354e87b6639c1bbe043f.tar.gz"},
	}

	for _, test := range tests {
		request := ArchiveRequest{
			relativeRepoPath: test.relativeRepoPath,
			commit:           test.commit,
			prefix:           test.prefix,
			format:           test.format,
		}

		hash := request.Hash()

		if hash != test.expectedHash {
			t.Errorf("expected \"%v\", got \"%v\"", test.expectedHash, hash)
		}
	}
}

func TestArchiveRequest_Execute(t *testing.T) {
	var tests = []struct {
		archiveError error
		expectedPath string
	}{
		{nil, "/archives/c8c0e5b45c70bf627f332b6488c651859c7d390d.zip"},
		{errors.New("oops"), ""},
	}

	for _, test := range tests {
		repository := &testRepository{archiveError: test.archiveError}

		request := ArchiveRequest{
			repository:       repository,
			relativeRepoPath: "ab.git",
			commit:           "cafebabe",
			prefix:           "prefix/",
			format:           "zip",
			archiveDir:       "/archives",
		}

		request.Execute()
		result := request.Result()
		archiveResult := result.(*ArchiveResult)

		if archiveResult.Error != test.archiveError {
			t.Errorf("expected %v, got %v", test.archiveError, archiveResult.Error)
		}

		if archiveResult.Path != test.expectedPath {
			t.Errorf("expected \"%v\", got \"%v\"", test.expectedPath, archiveResult.Path)
		}
	}
}
