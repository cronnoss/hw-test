package hw05parallelexecution

import (
	"errors"
	"sync"
)

var (
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	ErrInvalidWorkersCount = errors.New("invalid workers count")
)

type Task func() error

// Counter is a type that uses a mutex to allow safe concurrent increments.
type Counter struct {
	m     sync.Mutex
	value int
}

// Increment safely increments the counter's value.
func (c *Counter) Increment() {
	c.m.Lock()
	c.value++
	c.m.Unlock()
}

// Value safely gets the counter's value.
func (c *Counter) Value() int {
	c.m.Lock()
	defer c.m.Unlock()
	return c.value
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrInvalidWorkersCount
	}

	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	counter := Counter{}
	queue := make(chan Task, len(tasks))
	wg := sync.WaitGroup{}

	go func() {
		for _, task := range tasks {
			queue <- task
		}
		close(queue)
	}()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if task, ok := <-queue; ok && counter.Value() < m {
					if res := task(); res != nil {
						counter.Increment()
					}
				} else {
					return
				}
			}
		}()
	}

	wg.Wait()

	if counter.Value() >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
