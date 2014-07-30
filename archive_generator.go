package main

import (
	"crypto/sha1"
	"fmt"
)

type ArchiveGenerator struct {
	repositoryStore RepositoryStore
	requestQueue    chan *ArchiveRequest
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

	request := g.newArchiveRequest(repository.Path(), commit, prefix, format)
	g.requestQueue <- request
	result := <-request.ResultChan

	return result.Path, result.Error
}

func (g *ArchiveGenerator) newArchiveRequest(repoPath, commit, prefix, format string) *ArchiveRequest {
	return NewArchiveRequest(NewArchiveJob(repoPath, commit, prefix, format))
}

type ArchiveRequest struct {
	Job        *ArchiveJob
	ResultChan chan *ArchiveResult
}

func NewArchiveRequest(job *ArchiveJob) *ArchiveRequest {
	return &ArchiveRequest{
		Job:        job,
		ResultChan: make(chan *ArchiveResult),
	}
}

type ArchiveJob struct {
	RepoPath string
	Commit   string
	Prefix   string
	Format   string
	Filename string
	Result   *ArchiveResult
}

func NewArchiveJob(repoPath, commit, prefix, format string) *ArchiveJob {
	job := &ArchiveJob{
		RepoPath: repoPath,
		Commit:   commit,
		Prefix:   prefix,
		Format:   format,
	}
	data := []byte(repoPath + commit + prefix)
	job.Filename = fmt.Sprintf("%x.%v", sha1.Sum(data), format)

	return job
}

type ArchiveResult struct {
	Path  string
	Error error
}
