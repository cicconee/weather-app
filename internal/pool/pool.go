package pool

type Pool struct {
	workers int
	jobCh   chan func()
}

func New(workerCount int, jobChanSize int) *Pool {
	return &Pool{
		workers: workerCount,
		jobCh:   make(chan func(), jobChanSize),
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workers; i++ {
		go func() {
			for job := range p.jobCh {
				job()
			}
		}()
	}
}

func (p *Pool) Add(f func()) {
	p.jobCh <- f
}
