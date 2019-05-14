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
	//пробовала распараллеливать след. образом, изменяя количество воркеров, время выполнения только увеличиватеся
	//wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	//for i := 0; i < 1; i++ {
	//	wg.Add(1)
	//	fmt.Println("SingleHash, распараллеливаем вычесления по воркерам, воркер", i)
	//	go func(in, out chan interface{}, wg *sync.WaitGroup) {
	//		defer wg.Done()
	for input := range in {
		// Содержимое канала в строку
		data := fmt.Sprint(input)
		fmt.Println(data, "SingleHash data ", data)
		mu.Lock()
		sinderCrc := DataSignerCrc32(data)
		mu.Unlock()
		singerMd5 := DataSignerMd5(data)
		mu.Lock()
		singerCrcMd5 := DataSignerCrc32(singerMd5)
		mu.Unlock()
		// fmt.Println(data, "SingleHash md5(data)", singerMd5)
		// fmt.Println(data, "SingleHash crc32(md5(data)", singerCrcMd5)
		// fmt.Println(data, "SingleHash crc32(data)", sinderCrc)

		singerResult := fmt.Sprintf("%v~%v", sinderCrc, singerCrcMd5)

		fmt.Println(data, "SingleHash result", singerResult)

		// fmt.Println(data, "SingleHash запись в in", singerResult)
		out <- singerResult
		//runtime.Gosched() // отдаем право брать задачу другому воркеру
	}
	//	}(in, out, wg)
	//}
	//wg.Wait()

}

/*
	MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
	где th=0..5 ( т.е. 6 хешей на каждое входящее значение ),
	потом берёт конкатенацию результатов в порядке расчета (0..5),
	где data - то что пришло на вход (и ушло на выход из SingleHash)
*/
func MultiHash(in, out chan interface{}) {

	//пробовала распараллеливать след. образом, изменяя количество воркеров, время выполнения только увеличиватеся
	//wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	//for i := 0; i < 1; i++ {
	//	wg.Add(1)
	//	fmt.Println("MultiHash, распараллеливаем вычесления по воркерам, воркер", i)
	//	go func(in, out chan interface{}, wg *sync.WaitGroup) {
	//		defer wg.Done()
	for input := range in {

		// Содержимое канала в строку
		data := fmt.Sprint(input)
		// fmt.Println(data, "MultiHash data ", data)

		tmp := ""
		for i := 0; i < 6; i++ {
			mu.Lock()
			tmp = tmp + DataSignerCrc32(fmt.Sprint(i)+data)
			mu.Unlock()
		}

		fmt.Println(data, "MultiHash result", tmp)
		out <- tmp
		//runtime.Gosched()
	}

	//	}(in, out, wg)
	//}
	//wg.Wait()

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

	chIn := make(chan interface{}, 5)
	chOut := make(chan interface{}, 5)

	for _, currentJob := range jobs {
		wg.Add(1)

		//Входящий поток - это исходящий из предыдущего шага
		chIn = chOut
		chOut = make(chan interface{}, 10)

		go func(currentJob job, chIn, chOut chan interface{}, wg *sync.WaitGroup) {
			defer func(chOut chan interface{}, wg *sync.WaitGroup) {
				wg.Done()
				close(chOut)

			}(chOut, wg)

			currentJob(chIn, chOut)
		}(currentJob, chIn, chOut, wg)
	}

	wg.Wait()
}

func main() {
}
