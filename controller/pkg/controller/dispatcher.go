package controller

const numScanners = 5

var (
	MaxWorker = numScanners
	MaxQueue  = numScanners
)

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	jobQueue   chan Job
	workerPool chan chan Job
	maxWorkers int

}

func NewDispatcher(jobQueue chan Job, maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher {
		jobQueue:   jobQueue,
		workerPool: pool,
		maxWorkers: maxWorkers,
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(i + 1, d.workerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
			case job := <-d.jobQueue:
				// a job request has been received
				go func(job Job) {

					// try to obtain a worker job channel that is available.
					// this will block until a worker is idle
					workerJobQueue:= <-d.workerPool

					// dispatch the job to the worker job channel
					workerJobQueue <- job
				}(job)
		}
	}
}

