package main

import "log"

func ArchiveWorker(jobs <-chan *ArchiveJob, completedJobs chan<- *ArchiveJob, tmpDir string) {
	for job := range jobs {
		log.Printf("processing job %v...", job)

		outPath := tmpDir + "/" + job.Filename
		repository := &GitRepository{job.RepoPath}

		if err := repository.Archive(job.Commit, job.Prefix, job.Format, outPath); err != nil {
			job.Result = &ArchiveResult{"", err}
		} else {
			job.Result = &ArchiveResult{outPath, nil}
		}

		log.Printf("finished job %v", job)
		completedJobs <- job
	}
}
