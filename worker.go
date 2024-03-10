package main

import (
	"context"
	"errors"
	"sync"

	"github.com/pbergman/logger"
)

type Job interface {
	Run(ctx context.Context, logger *logger.Logger)
}

type WorkerState = uint8

const (
	WorkerStateClosed WorkerState = iota
	WorkerStateOpen
	WorkerStateStarted
)

func NewWorker(logger *logger.Logger) *Worker {
	return &Worker{
		queue:  make(chan Job, 15),
		logger: logger,
		wg:     new(sync.WaitGroup),
		state:  WorkerStateOpen,
	}
}

type Worker struct {
	queue  chan Job
	logger *logger.Logger
	wg     *sync.WaitGroup
	state  WorkerState
}

func (w *Worker) Queue(j Job) error {
	if w.state == WorkerStateClosed {
		return errors.New("cannot queue job because worker is closed")
	}
	w.queue <- j
	return nil
}

func (w *Worker) Wait() {

	if w.state != WorkerStateStarted {
		return
	}

	w.wg.Wait()
}

func (w *Worker) Close() error {
	if w.state == WorkerStateClosed {
		return nil
	}
	close(w.queue)
	w.state = WorkerStateClosed
	return nil
}

func (w *Worker) Start(cxt context.Context, max uint8) {
	w.state = WorkerStateStarted
	w.wg.Add(int(max))

	for i := uint8(0); i < max; i++ {

		go func() {
			defer w.wg.Done()

			for job := range w.queue {
				job.Run(cxt, w.logger)
			}
		}()
	}
}

//func startWorker(cxt context.Context, max uint8, queue chan Job, wg *sync.WaitGroup, logger *logger.Logger) {
//
//	wg.Add(int(max))
//
//	for i := uint8(0); i < max; i++ {
//
//		go func() {
//			defer wg.Done()
//
//			for job := range queue {
//				job.Run(cxt, logger)
//			}
//		}()
//	}
//}
