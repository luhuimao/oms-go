package engine

type Dispatcher struct {
	workers []*Worker
}

func NewDispatcher(n int) *Dispatcher {
	ws := make([]*Worker, n)
	for i := range ws {
		ws[i] = NewWorker()
	}
	return &Dispatcher{workers: ws}
}

func (d *Dispatcher) Dispatch(key int64, fn func()) {
	idx := key % int64(len(d.workers))
	d.workers[idx].Submit(fn)
}
