package main

import "log"

func ArchiveWorker(jobs <-chan *ArchiveJob, completedJobs chan<- *ArchiveJob, tmpDir string) {
	for job := range jobs {
		log.Printf("worker: processing job %v...", job)

		outPath := tmpDir + "/" + job.Filename
		repository := &GitRepository{job.RepoPath}

		if err := repository.Archive(job.Commit, job.Prefix, job.Format, outPath); err != nil {
			job.Result = &ArchiveResult{"", err}
		} else {
			job.Result = &ArchiveResult{outPath, nil}
		}

		log.Printf("worker: finished job %v", job)
		completedJobs <- job
	}
}

func ArchiveWorkerPool(num int, tmpDir string) (chan *ArchiveJob, chan *ArchiveJob) {
	jobs := make(chan *ArchiveJob)
	results := make(chan *ArchiveJob)

	for n := 0; n < num; n++ {
		go ArchiveWorker(jobs, results, tmpDir)
	}

	return jobs, results
}
