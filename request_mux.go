package main

func RequestMux(requests chan *ArchiveRequest, jobs chan *ArchiveJob, results chan *ArchiveJob) {
	queues := make(map[string][]*ArchiveRequest)

	for {
		select {
		case request := <-requests:
			job := request.Job
			queues[job.Filename] = append(queues[job.Filename], request)

			if len(queues[job.Filename]) == 1 {
				go func() {
					jobs <- job
				}()
			}
		case job := <-results:
			for _, request := range queues[job.Filename] {
				request.ResultChan <- job.Result
			}

			delete(queues, job.Filename)
		}
	}
}
