package engine

type Worker struct {
	ch chan func()
}

func NewWorker() *Worker {
	w := &Worker{ch: make(chan func(), 1024)}
	go func() {
		for fn := range w.ch {
			fn()
		}
	}()
	return w
}

func (w *Worker) Submit(fn func()) {
	w.ch <- fn
}
