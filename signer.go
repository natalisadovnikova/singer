package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
)

/*
	считает значение crc32(data)+"~"+crc32(md5(data)) ( конкатенация двух строк через ~),
	где data - то что пришло на вход (по сути - числа из первой функции
*/
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		input := i
		data := fmt.Sprint(input)
		singerMd5 := DataSignerMd5(data)

		wg.Add(1)
		go func() {
			defer wg.Done()
			ch1 := make(chan interface{})
			ch2 := make(chan interface{})
			go func() {
				ch1 <- DataSignerCrc32(singerMd5)
				close(ch1)
			}()
			go func() {
				ch2 <- DataSignerCrc32(data)
				close(ch2)
			}()
			singerResult := fmt.Sprintf("%v~%v", <-ch2, <-ch1)
			out <- singerResult
		}()
	}
	wg.Wait()
}

/*
	MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
	где th=0..5 ( т.е. 6 хешей на каждое входящее значение ),
	потом берёт конкатенацию результатов в порядке расчета (0..5),
	где data - то что пришло на вход (и ушло на выход из SingleHash)
*/
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		data := fmt.Sprint(i)
		wg.Add(1)
		go func() {
			defer wg.Done()

			tmp := ""
			ch1 := make(chan map[int]string)

			for i := 0; i < 6; i++ {
				wg.Add(1)

				iter := i
				go func() {
					defer wg.Done()
					ch1 <- map[int]string{iter: DataSignerCrc32(fmt.Sprint(iter) + data)}
				}()
			}

			counter := 0
			m := make(map[int]string)
			for SignerCrc32Map := range ch1 {
				counter++
				for key1, val1 := range SignerCrc32Map {
					m[key1] = val1
				}
				if counter > 5 {
					close(ch1)
				}
			}
			var keys []int
			for k := range m {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			for _, k := range keys {
				tmp = tmp + m[k]
			}

			out <- tmp
		}()
	}
	wg.Wait()
}

/*
	CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/),
	объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
*/
func CombineResults(in, out chan interface{}) {
	buf := make([]string, 0)
	for input := range in {
		data := fmt.Sprint(input)
		buf = append(buf, data)
	}
	sort.Slice(buf, func(i, j int) bool {
		return buf[i] < buf[j]
	})
	tmp := strings.Join(buf, "_")
	out <- tmp
}

/*
Конвеерная обработка функций
*/
func ExecutePipeline(jobs ...job) {
	runtime.GOMAXPROCS(0)
	wg := &sync.WaitGroup{}

	chOut := make(chan interface{}, 5)

	for _, j := range jobs {
		wg.Add(1)

		//Входящий поток - это исходящий из предыдущего шага
		chIn := chOut
		chOut = make(chan interface{}, 10)
		currentJob := j

		go func(chOut chan interface{}) {
			defer close(chOut)
			defer wg.Done()
			currentJob(chIn, chOut)
		}(chOut)
	}

	wg.Wait()
}

func main() {
}
