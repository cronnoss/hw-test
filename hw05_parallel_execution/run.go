package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	ErrInvalidWorkersCount = errors.New("invalid workers count")
)

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrInvalidWorkersCount
	}

	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	var errCounter int32
	queue := make(chan Task)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go increment(queue, &errCounter, &wg)
	}
	checkErrorLimits := m > 0
	for _, task := range tasks {
		if checkErrorLimits && atomic.LoadInt32(&errCounter) >= int32(m) {
			break
		}
		queue <- task
	}
	close(queue)

	wg.Wait()

	if errCounter > 0 {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func increment(queue chan Task, errCounter *int32, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range queue {
		err := task()
		if err != nil {
			atomic.AddInt32(errCounter, 1)
		}
	}
}
