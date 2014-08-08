package main

import "log"

func RequestMux(jobs chan *ArchiveJob, results chan *ArchiveJob) chan *ArchiveRequest {
	requests := make(chan *ArchiveRequest)

	go func() {
		queues := make(map[string][]*ArchiveRequest)

		for {
			select {
			case request := <-requests:
				job := request.Job
				log.Printf("mux: appending request %v to queue for %v", request, job.Filename)
				queues[job.Filename] = append(queues[job.Filename], request)

				if len(queues[job.Filename]) == 1 {
					log.Printf("mux: scheduling job %v", job)
					go func() {
						jobs <- job
					}()
				}
			case job := <-results:
				log.Printf("mux: sending job result to requests from queue for %v", job.Filename)
				for _, request := range queues[job.Filename] {
					request.ResultChan <- job.Result
				}

				delete(queues, job.Filename)
			}
		}
	}()

	return requests
}
