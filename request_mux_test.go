package main

import (
	"testing"
	"time"
)

var jobCompletionChans = make(map[*ArchiveJob]chan struct{})

func newTestRequest(filename string) (*ArchiveRequest, chan struct{}) {
	job := &ArchiveJob{Filename: filename}
	ch := make(chan struct{})
	jobCompletionChans[job] = ch

	return NewArchiveRequest(job), ch
}

func testWorker(jobs <-chan *ArchiveJob, results chan<- *ArchiveJob) {
	for job := range jobs {
		<-jobCompletionChans[job] // block on chan read, closing this channel signals job completion
		job.Result = &ArchiveResult{Path: "/the/path/" + job.Filename}
		results <- job
	}
}

func TestRequestMux(t *testing.T) {
	jobs := make(chan *ArchiveJob)
	completedJobs := make(chan *ArchiveJob)

	go testWorker(jobs, completedJobs)
	go testWorker(jobs, completedJobs)

	requests := RequestMux(jobs, completedJobs)

	// Send 6 requests (having 4 unique jobs).
	//
	// None of these should block, regardles of the number of workers
	// processing the jobs queue.
	// We want to handle unlimited number of requests, groupping them together
	// by job's filename, sending them job's result when it's ready.

	request1, jobCompletionChan1 := newTestRequest("file-a")
	requests <- request1
	request2, jobCompletionChan2 := newTestRequest("file-b")
	requests <- request2
	request3, _ := newTestRequest("file-a")
	requests <- request3
	request4, _ := newTestRequest("file-b")
	requests <- request4
	request5, jobCompletionChan5 := newTestRequest("file-c")
	requests <- request5
	request6, _ := newTestRequest("file-d")
	requests <- request6

	// Finish job for request 2. This should return result for
	// request 2 and 4 as they have the same filename ("file-b").

	close(jobCompletionChan2)

	result := <-request2.ResultChan
	if result.Path != "/the/path/file-b" {
		t.Errorf(`expected "/the/path/file-b", got %v`, result.Path)
	}

	result = <-request4.ResultChan
	if result.Path != "/the/path/file-b" {
		t.Errorf(`expected "/the/path/file-b", got %v`, result.Path)
	}

	// Finish job for request 1. This should return result for
	// request 1 and 3 as they have the same filename ("file-a").

	close(jobCompletionChan1)

	result = <-request1.ResultChan
	if result.Path != "/the/path/file-a" {
		t.Errorf(`expected "/the/path/file-a", got %v`, result.Path)
	}

	result = <-request3.ResultChan
	if result.Path != "/the/path/file-a" {
		t.Errorf(`expected "/the/path/file-a", got %v`, result.Path)
	}

	// Finish job for request 5. This should return result for
	// request 5 only as its the only one having filename "file-c".

	close(jobCompletionChan5)

	result = <-request5.ResultChan
	if result.Path != "/the/path/file-c" {
		t.Errorf(`expected "/the/path/file-c", got %v`, result.Path)
	}

	// Check that request 6 hasn't been completed yet.
	// Reading from it's result chan should be blocking.

	select {
	case <-request6.ResultChan:
		t.Errorf("request6 shouldn't get any response yet")
	case <-time.After(time.Millisecond * 1):
		// we should get here
	}
}
