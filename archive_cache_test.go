package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

func TestArchiveCache(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "")
	workDir := tmpDir + "/work"
	os.Mkdir(workDir, 0755)
	cacheDir := tmpDir + "/cache"
	os.Mkdir(cacheDir, 0755)

	actualJobs := make(chan *ArchiveJob)
	actualResults := make(chan *ArchiveJob)

	cachedJobs, cachedResults := ArchiveCache(actualJobs, actualResults, cacheDir)

	// put file1 in cache dir
	ioutil.WriteFile(cacheDir+"/file1", []byte("file1"), 0644)

	// enqueue job for file1
	job1 := &ArchiveJob{Filename: "file1"}
	cachedJobs <- job1

	// read job result - it should be ready
	completedJob := <-cachedResults
	if completedJob != job1 {
		t.Errorf("expected job %v to be cached", job1)
	}

	// confirm that it's the right file
	contents, _ := ioutil.ReadFile(completedJob.Result.Path)
	if string(contents) != "file1" {
		t.Errorf("expected file to have contents = \"file1\"")
	}

	// enqueue job for file2, which isn't cached
	job2 := &ArchiveJob{Filename: "file2"}
	cachedJobs <- job2

	// confirm that actual job was scheduled
	scheduledJob := <-actualJobs
	if scheduledJob != job2 {
		t.Errorf("expected job %v to be scheduled", job2)
	}

	// complete actual job
	ioutil.WriteFile(workDir+"/a-file", []byte("file2"), 0644)
	job2.Result = &ArchiveResult{workDir + "/a-file", nil}
	actualResults <- job2

	// read job result - it should be ready
	completedJob = <-cachedResults
	if completedJob != job2 {
		t.Errorf("expected job %v to be completed", job2)
	}

	// confirm that it's the right file
	contents, _ = ioutil.ReadFile(completedJob.Result.Path)
	if string(contents) != "file2" {
		t.Errorf("expected file to have contents = \"file2\"")
	}

	// confirm that the file was moved from work dir
	if _, err := os.Stat(workDir + "/a-file"); err == nil {
		t.Errorf("expected a-file to be moved from original dir")
	}

	// enqueue job for file3, which isn't cached
	job3 := &ArchiveJob{Filename: "file3"}
	cachedJobs <- job3

	// confirm that actual job was scheduled
	scheduledJob = <-actualJobs
	if scheduledJob != job3 {
		t.Errorf("expected job %v to be scheduled", job3)
	}

	// fail actual job
	jobError := errors.New("oops")
	job3.Result = &ArchiveResult{"", jobError}
	actualResults <- job3

	// read job result - it should be ready
	completedJob = <-cachedResults
	if completedJob != job3 {
		t.Errorf("expected job %v to be completed", job3)
	}

	// confirm that the result has error
	if completedJob.Result.Error != jobError {
		t.Errorf("expected job to fail with error %v", jobError)
	}

	// clean up
	os.RemoveAll(tmpDir)
}
