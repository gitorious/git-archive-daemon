package main

import (
	"fmt"
	"log"
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
					log.Println("cache: hit for %v", job)
					job.Result = &ArchiveResult{cachedPath, nil}
					results <- job
				} else {
					log.Println("cache: miss for %v", job)
					upstreamJobs <- job
				}

			case job := <-upstreamResults:
				result := job.Result
				if result.Error == nil {
					log.Println("cache: saving result for %v", job)
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
