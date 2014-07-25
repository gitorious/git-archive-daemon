package main

import (
	"testing"
	"time"
)

type testRequest struct {
	text         string
	result       string
	resultChan   chan interface{}
	completeChan chan struct{}
}

func newTestRequest(text string) *testRequest {
	return &testRequest{
		text:         text,
		resultChan:   make(chan interface{}),
		completeChan: make(chan struct{}),
	}
}

func (r *testRequest) Job() Job {
	return r
}

func (r *testRequest) ResultChan() chan interface{} {
	return r.resultChan
}

func (r *testRequest) Hash() string {
	return r.text
}

func (r *testRequest) Execute() {
	<-r.completeChan
	r.result = r.text + r.text
}

func (r *testRequest) Result() interface{} {
	return r.result
}

func (r *testRequest) completeJob() {
	// Close the channel. This signals completion of the job to Execute().
	close(r.completeChan)
}

func TestRequestMux(t *testing.T) {
	requests := make(chan Request)
	jobs := make(chan Job)
	completedJobs := make(chan Job)

	for i := 0; i < 2; i++ {
		go func() {
			for job := range jobs {
				job.Execute()
				completedJobs <- job
			}
		}()
	}

	go RequestMux(requests, jobs, completedJobs)

	// Send 6 requests (having 4 unique jobs).
	//
	// None of these should block, regardles of the number of workers
	// processing the jobs queue.
	// We want to handle unlimited number of requests, groupping them together
	// by job's unique hash, sending them job's result when it's ready.

	request1 := newTestRequest("a")
	requests <- request1
	request2 := newTestRequest("b")
	requests <- request2
	request3 := newTestRequest("a")
	requests <- request3
	request4 := newTestRequest("b")
	requests <- request4
	request5 := newTestRequest("c")
	requests <- request5
	request6 := newTestRequest("d")
	requests <- request6

	// Finish response generation for request 2. This should return result for
	// request 2 and 4 as they have the same hash value ("b").

	request2.completeJob()

	result := <-request2.ResultChan()
	if result != "bb" {
		t.Errorf("expected bb, got %v", result)
	}

	result = <-request4.ResultChan()
	if result != "bb" {
		t.Errorf("expected bb, got %v", result)
	}

	// Finish response generation for request 1. This should return result for
	// request 1 and 3 as they have the same hash value ("a").

	request1.completeJob()

	result = <-request1.ResultChan()
	if result != "aa" {
		t.Errorf("expected aa, got %v", result)
	}

	result = <-request3.ResultChan()
	if result != "aa" {
		t.Errorf("expected aa, got %v", result)
	}

	// Finish response generation for request 5. This should return result for
	// request 5 only as its the only one having hash value of "c".

	request5.completeJob()

	result = <-request5.ResultChan()
	if result != "cc" {
		t.Errorf("expected cc, got %v", result)
	}

	// Check that request 6 hasn't been completed yet.
	// Reading from it's result chan should be blocking.

	select {
	case <-request6.ResultChan():
		t.Errorf("request6 shouldn't get any response yet")
	case <-time.After(time.Millisecond * 1):
		// we should get here
	}
}
