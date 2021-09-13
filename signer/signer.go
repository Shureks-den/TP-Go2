package main

import (
	"sync"
)

// сюда писать код

func ExecutePipeline(jobs ...job) {
	var waitgr sync.WaitGroup
	in := make(chan interface{})

	for _, curJob := range jobs {
		waitgr.Add(1)
		out := make(chan interface{})
		go func(waitgroup *sync.WaitGroup, pipelineJob job, in, out chan interface{}) {
			defer waitgroup.Done()
			pipelineJob(in, out)
			close(out)
		}(&waitgr, curJob, in, out)
		in = out
	}

	waitgr.Wait()
}
