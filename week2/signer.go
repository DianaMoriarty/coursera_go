package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 100)
	for _, oneJob := range jobs {
		out := make(chan interface{}, 100)
		wg.Add(1)
		go func(jobFunc job, in chan interface{}, out chan interface{}) {
			defer close(out)
			defer wg.Done()
			jobFunc(in, out)
		}(oneJob, in, out)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for data := range in {
		val := strconv.Itoa(data.(int))
		wg.Add(1)
		go func(data string) {
			defer wg.Done()
			firstHash, secondHash := make(chan string), make(chan string)
			go func() {
				firstHash <- DataSignerCrc32(data)
			}()
			go func() {
				mu.Lock()
				md5data := DataSignerMd5(data)
				mu.Unlock()
				secondHash <- DataSignerCrc32(md5data)
			}()
			out <- <-firstHash + "~" + <-secondHash
		}(val)
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		dataString := data.(string)
		go func(in string) {
			var arr [6]chan string
			for i := 0; i <= 5; i++ {
				arr[i] = make(chan string)
				go func(i int, str string) {
					arr[i] <- DataSignerCrc32(strconv.Itoa(i) + str)
				}(i, in)
			}
			result := ""
			//result := strings.Join(arr[:], "")
			for i := 0; i <= 5; i++ {
				result += <-arr[i]
			}
			out <- result
			wg.Done()
		}(dataString)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var s []string
	for data := range in {
		s = append(s, data.(string))
	}
	sort.Strings(s)
	out <- strings.Join(s, "_")
}

func main() {

}
