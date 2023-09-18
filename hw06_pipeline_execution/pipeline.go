package hw06pipelineexecution

import "time"

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	for _, stage := range stages {
		if stage != nil {
			in = stage(workStage(done, in))
		}
	}
	return in
}

func workStage(done In, in In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				if v != nil {
					select {
					case out <- v:
					case <-done:
						return
					}
				}
			default:
				time.Sleep(time.Microsecond)
			}
		}
	}()
	return out
}
