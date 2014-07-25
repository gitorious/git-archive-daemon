package main

type Request interface {
	Job() Job
	ResultChan() chan interface{}
}

type Job interface {
	Hash() string
	Execute()
	Result() interface{}
}

func RequestMux(requests chan Request, jobs chan Job, results chan Job) {
	queues := make(map[string][]Request)

	for {
		select {
		case request := <-requests:
			job := request.Job()
			hash := job.Hash()
			queues[hash] = append(queues[hash], request)

			if len(queues[hash]) == 1 {
				go func() {
					jobs <- job
				}()
			}
		case job := <-results:
			hash := job.Hash()

			for _, request := range queues[hash] {
				request.ResultChan() <- job.Result()
			}

			delete(queues, hash)
		}
	}
}
