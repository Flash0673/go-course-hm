package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		out := make(chan interface{})
		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			j(in, out)
		}(j, in, out)

		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for val := range in {
		val := fmt.Sprintf("%v", val)
		md5 := DataSignerMd5(val)

		wg.Add(1)
		go func(val string, md5 string) {
			defer wg.Done()
			hash_1 := HashFunc(val)
			hash_2 := HashFunc(md5)

			out <- fmt.Sprintf("%s~%s", <-hash_1, <-hash_2)
		}(val, md5)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for val := range in {
		wg.Add(1)
		val := fmt.Sprint(val)

		go func() {	
			defer wg.Done()
			chs := make([]<-chan string, 6)

			for i := 0; i < 6; i++ {
				
				ch := HashFunc(strconv.Itoa(i) + val)

				mu.Lock()
				chs[i] = ch
				mu.Unlock()
			}

			res := make([]string, 6)
			for i, c := range chs {
				res[i] = <- c
			}

			out <- strings.Join(res, "")
		}()
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results = []string{}
	for x := range in {
		results = append(results, x.(string))
	}

	sort.Strings(results)
	out <- strings.Join(results, "_")
}

func HashFunc(val string) <-chan string {
	out := make(chan string)
	go func(val string) {
		defer close(out)
		out <- DataSignerCrc32(val)
	}(val)

	return out
}

func Md5Hash(val string) <-chan string {
	out := make(chan string)
	go func(val string) {
		defer close(out)
		out <- DataSignerMd5(val)
	}(val)

	return out
}