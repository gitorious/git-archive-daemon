package main

import (
	"fmt"
	"os"
)

func ArchiveCache(
	upstreamJobs chan *ArchiveJob,
	upstreamResults chan *ArchiveJob,
	cachePath string,
) (chan *ArchiveJob, chan *ArchiveJob) {

	jobs := make(chan *ArchiveJob)
	results := make(chan *ArchiveJob)

	go func() {
		for {
			select {
			case job := <-jobs:
				cachedPath := fmt.Sprintf("%v/%v", cachePath, job.Filename)
				if _, err := os.Stat(cachedPath); err == nil {
					job.Result = &ArchiveResult{cachedPath, nil}
					results <- job
				} else {
					upstreamJobs <- job
				}

			case job := <-upstreamResults:
				result := job.Result
				if result.Error == nil {
					cachedPath := fmt.Sprintf("%v/%v", cachePath, job.Filename)
					os.Rename(result.Path, cachedPath)
					job.Result = &ArchiveResult{cachedPath, nil}
				}
				results <- job
			}
		}
	}()

	return jobs, results
}
