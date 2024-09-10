package main

import (
	"sort"
	"strings"
	"sync"
	"fmt"
	"strconv"
)

func main() {
	// start := time.Now()
	in, out := make(chan interface{}), make(chan interface{})
	go SingleHash(in, out)
	in <- "1"
	fmt.Println(<-out)

	// fmt.Println(DataSignerCrc32("hello") + "~" + DataSignerCrc32(DataSignerMd5("hello")))
	// fmt.Println(time.Now().Sub(start))
}

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
	fmt.Println("Single hash entered")
	for x := range in {
		resChan := make(chan string)
		md5Chan := make(chan string)

		go func(x interface{}) {
			resChan <- DataSignerCrc32(fmt.Sprintf("%v", x))
		}(x)

		go func(x interface{}) {
			defer close(md5Chan)
			md5Chan <- DataSignerMd5(fmt.Sprintf("%v", x))
		}(x)

		go func() {
			resChan <- DataSignerCrc32(<- md5Chan)
		}()



		res := make([]string, 2)
		for i := 0; i<2; i++ {
			s := <- resChan
			res[i] = s
		}
		close(resChan)
		out <- strings.Join(res, "~")
		fmt.Println("Single done")
	}
	
}

func MultiHash(in, out chan interface{}) {
	fmt.Println("Multi hash entered")
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
		for data := range in {
		res := make([]string, 6)

		for i := range res {
			wg.Add(1)
			go func(th int) {
				defer wg.Done()
				hash := DataSignerCrc32(strconv.Itoa(th) + fmt.Sprintf("%v", data))

				mu.Lock()
				res[th] = hash
				mu.Unlock()
			}(i)
		}
		wg.Wait()
		out <- strings.Join(res, "")
	} 

	// mu := &sync.Mutex{}
	// for data := range in {
	// 	chans := make([]chan string, 6)
	// 	for i := range chans {
	// 		chans[i] = make(chan string)
	// 	}

	// 	for i, c := range chans {
	// 		go func(th int, c chan string) {
	// 			c <- DataSignerCrc32(strconv.Itoa(th) + data.(string))
	// 		}(i, c)
	// 	}
	// 	res := ""
	// 	wg := &sync.WaitGroup{}
	// 	for _, c := range chans {
	// 		wg.Add(1)
	// 		go func(c chan string) {
	// 			defer close(c)
	// 			defer wg.Done()
	// 			mu.Lock()
	// 			res += <- c
	// 			mu.Unlock()
	// 		}(c)
	// 	}
	// 	wg.Wait()
	// 	out <- res
	// }
	fmt.Println("Multi hash done")
}

func CombineResults(in, out chan interface{}) {
	fmt.Println("Combine results entered")
	var results = []string{}
	for x := range in {
		results = append(results, x.(string))
	}

	sort.Strings(results)
	out <- strings.Join(results, "_")
	fmt.Println("Combine results done")
}
