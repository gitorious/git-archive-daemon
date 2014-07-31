package main

import (
	"fmt"
	"io"
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
					if err := moveFile(result.Path, cachedPath); err != nil {
						job.Result = &ArchiveResult{"", err}
					} else {
						job.Result = &ArchiveResult{cachedPath, nil}
					}
				}
				results <- job
			}
		}
	}()

	return jobs, results
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		if err = copyFile(src, dst); err != nil {
			return err
		}

		if err = os.Remove(src); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)

	return err
}
