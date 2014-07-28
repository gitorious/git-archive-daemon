package main

import (
	"crypto/sha1"
	"fmt"
)

type ArchiveGenerator struct {
	repositoryStore RepositoryStore
	archiveDir      string
	requestQueue    chan Request
}

func (g *ArchiveGenerator) GenerateArchive(path, ref, prefix, format string) (string, error) {
	repository, err := g.repositoryStore.GetRepository(path)
	if err != nil {
		return "", REPOSITORY_NOT_FOUND
	}

	commit, err := repository.ResolveRef(ref)
	if err != nil {
		return "", REF_NOT_FOUND
	}

	request := g.newArchiveRequest(repository, path, commit, prefix, format)
	g.requestQueue <- request

	result := <-request.ResultChan()
	archiveResult := result.(*ArchiveResult)

	return archiveResult.Path, archiveResult.Error
}

func (g *ArchiveGenerator) newArchiveRequest(repository Repository, relativeRepoPath, commit, prefix, format string) *ArchiveRequest {
	return &ArchiveRequest{
		repository:       repository,
		relativeRepoPath: relativeRepoPath,
		commit:           commit,
		prefix:           prefix,
		format:           format,
		archiveDir:       g.archiveDir,
		resultChan:       make(chan interface{}),
	}
}

type ArchiveRequest struct {
	repository       Repository
	relativeRepoPath string
	commit           string
	prefix           string
	format           string
	archiveDir       string
	result           *ArchiveResult
	resultChan       chan interface{}
}

var _ Request = &ArchiveRequest{}
var _ Job = &ArchiveRequest{}

func (r *ArchiveRequest) Job() Job {
	return r
}

func (r *ArchiveRequest) ResultChan() chan interface{} {
	return r.resultChan
}

func (r *ArchiveRequest) Hash() string {
	data := []byte(r.relativeRepoPath + r.commit + r.prefix)
	return fmt.Sprintf("%x.%v", sha1.Sum(data), r.format)
}

func (r *ArchiveRequest) Execute() {
	outPath := r.archiveDir + "/" + r.Hash()

	if err := r.repository.Archive(r.commit, r.prefix, r.format, outPath); err != nil {
		r.result = &ArchiveResult{"", err}
	} else {
		r.result = &ArchiveResult{outPath, nil}
	}
}

func (r *ArchiveRequest) Result() interface{} {
	return r.result
}

type ArchiveResult struct {
	Path  string
	Error error
}
