package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// сюда писать код

var multiHashCounter int = 6

func ExecutePipeline(jobs ...job) {
	var waitgr sync.WaitGroup
	in := make(chan interface{})

	for _, curJob := range jobs {
		waitgr.Add(1)
		out := make(chan interface{})
		// выполнение работы в пайплайне
		go func(waitgroup *sync.WaitGroup, pipelineJob job, in, out chan interface{}) {
			defer waitgroup.Done()
			pipelineJob(in, out)
			close(out)
		}(&waitgr, curJob, in, out)
		// перенаправление каналов
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

			go calculateHash(src32, DataSignerCrc32, data)
			go calculateHash(src32md5, DataSignerCrc32, md5)

			src32Res := <-src32
			src32md5Res := <-src32md5
			out <- src32Res + "~" + src32md5Res
		}(&waitgr, data, md5hash, out)
	}
	waitgr.Wait()
}

// функция для уменьшения копипасты в singleHasher
func calculateHash(ch chan string, f func(string) string, data string) {
	ch <- f(data)
}

func MultiHash(in, out chan interface{}) {
	var waitgr sync.WaitGroup

	for i := range in {
		waitgr.Add(1)

		go func(waitgr *sync.WaitGroup, data interface{}, out chan interface{}) {
			defer waitgr.Done()
			var innerWaitgr sync.WaitGroup
			hashArray := make([]string, multiHashCounter)

			for index := 0; index < multiHashCounter; index++ {
				innerWaitgr.Add(1)
				hashData := fmt.Sprintf("%v%v", index, data)
				go multiHashItteration(&innerWaitgr, hashData, hashArray, index)
			}
			innerWaitgr.Wait()
			multiHash := strings.Join(hashArray, "")
			out <- multiHash
		}(&waitgr, i, out)
	}
	waitgr.Wait()
}

func multiHashItteration(wg *sync.WaitGroup, s string, array []string, index int) {
	defer wg.Done()
	crc32Hash := DataSignerCrc32(s)
	array[index] = crc32Hash
}

func CombineResults(in, out chan interface{}) {
	var resArray []string

	for i := range in {
		resArray = append(resArray, i.(string))
	}

	sort.Strings(resArray)
	joinedData := strings.Join(resArray, "_")
	out <- joinedData
}
