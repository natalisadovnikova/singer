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
	mu := &sync.Mutex{}

	for i := range in {
		data := fmt.Sprint(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Содержимое канала в строку

			fmt.Println(data, "SingleHash data ", data)
			mu.Lock()
			sinderCrc := DataSignerCrc32(data)
			mu.Unlock()
			singerMd5 := DataSignerMd5(data)
			mu.Lock()
			singerCrcMd5 := DataSignerCrc32(singerMd5)
			mu.Unlock()
			singerResult := fmt.Sprintf("%v~%v", sinderCrc, singerCrcMd5)

			fmt.Println(data, "SingleHash result", singerResult)
			out <- singerResult
		}()

		//runtime.Gosched() // отдаем право брать задачу другому воркеру
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
			for i := 0; i < 6; i++ {
				tmp = tmp + DataSignerCrc32(fmt.Sprint(i)+data)
			}
			fmt.Println(data, "MultiHash result", tmp)
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
	fmt.Println("CombineResults get data")
	buf := make([]string, 0)
	for input := range in {
		data := fmt.Sprint(input)
		buf = append(buf, data)
	}
	sort.Slice(buf, func(i, j int) bool {
		return buf[i] < buf[j]
	})
	tmp := strings.Join(buf, "_")

	fmt.Println("CombineResults result", tmp)
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
