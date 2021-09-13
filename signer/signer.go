package main

import (
	"fmt"
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

func SingleHash(in, out chan interface{}) {
	var waitgr sync.WaitGroup

	for i := range in {
		waitgr.Add(1)
		data := fmt.Sprintf("%v", i)
		md5hash := DataSignerMd5(data)

		go func(waitgroup *sync.WaitGroup, data, md5 string, out chan interface{}) {
			defer waitgroup.Done()

			src32 := make(chan string)
			src32md5 := make(chan string)

			go func(ch chan string, f func(string) string, data string) {
				ch <- f(data)
			}(src32, DataSignerCrc32, data)

			go func(ch chan string, f func(string) string, data string) {
				ch <- f(data)
			}(src32md5, DataSignerCrc32, md5)

			src32Res := <-src32
			src32md5Res := <-src32md5
			out <- src32Res + "~" + src32md5Res

		}(&waitgr, data, md5hash, out)
	}

	waitgr.Wait()
}
