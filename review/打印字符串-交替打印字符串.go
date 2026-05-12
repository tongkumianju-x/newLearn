package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// disorderPrint 无序打印a,b两个数组
func disorderPrint(a []int, b []string, size int) {

	var wg sync.WaitGroup
	alen := len(a)
	blen := len(b)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < alen; i += size {
			for j := 0; j < size; j++ {
				fmt.Print(a[i+j])
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < blen; i += size {
			for j := 0; j < size; j++ {
				fmt.Print(b[i+j])
			}
		}
	}()

	wg.Wait()

}

// atomicityPrint 原子性打印，谁先打印全打印
func atomicityPrint(a []int, b []string, size int) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	alen := len(a)
	blen := len(b)

	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		for i := 0; i < alen; i += size {
			for j := 0; j < size; j++ {
				fmt.Print(a[i+j])
			}
		}
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		for i := 0; i < blen; i += size {
			for j := 0; j < size; j++ {
				fmt.Print(b[i+j])
			}
		}
		mu.Unlock()
	}()

	wg.Wait()
}

// alternatePrint 采用缓冲区方式交替打印
func alternatePrint(a []int, b []string, size int) {
	aChannel := make(chan bool)
	bChannel := make(chan bool)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(a); i += size {
			//if len(a) > len(b) && i < len(b) {
			<-aChannel
			//}
			for j := 0; j < size; j++ {
				fmt.Print(a[i+j])
			}
			if len(a) > len(b) && i < len(b) {
				bChannel <- true
			}
		}
		close(aChannel)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(b); i += size {
			<-bChannel
			for j := 0; j < size; j++ {
				fmt.Print(b[i+j])
			}
			aChannel <- true
		}
		close(bChannel)
	}()

	aChannel <- true
	wg.Wait()

}

func main() {
	var a []int
	var b []string

	for i := 0; i < 26; i++ {
		a = append(a, i+1)
	}
	for i := 0; i < 26; i++ {
		b = append(b, string('A'+i))
	}
	for i := 0; i < 10; i++ {
		disorderPrint(a, b, 2)
		fmt.Println()
	}
	fmt.Print(strings.Repeat("-", 60))
	fmt.Println()
	for i := 0; i < 10; i++ {
		atomicityPrint(a, b, 2)
		fmt.Println()
	}
	fmt.Println(strings.Repeat("-", 60))
	for i := 0; i < 10; i++ {
		alternatePrint(a, b, 2)
		fmt.Println()
	}
}

// 交替打印a/b数组,交替间隔size--可无视长度的办法
func jiaotiPrint1(a []int, b []string, size int) {
	var wg sync.WaitGroup

	amutex := make(chan bool)
	bmutex := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(a); i = i + size {
			for j := 0; j < size && i+j < len(a); j++ {
				if len(a) > len(b) && i+j >= len(b) {
					time.Sleep(1 * time.Second)
					fmt.Print(a[i+j])
					fmt.Print("超额打印")
					continue
				}
				<-amutex
				fmt.Print(a[i+j])
				bmutex <- true
			}
		}
		close(amutex)

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(b); i = i + size {
			for j := 0; j < size && i+j < len(b); j++ {
				if len(b) > len(a) && i+j >= len(a) {
					fmt.Print(b[i+j])
					fmt.Print("超额打印")
					continue
				}
				<-bmutex
				fmt.Print(b[i+j])
				//if len(b) >= len(a) && i+j == len(a)-1 {
				//    continue
				//}
				if (len(b) >= len(a) && i+j == len(a)-1) || (len(b) < len(a) && i+j == len(b)-1) {
					fmt.Print("跳过")
					continue
				}
				amutex <- true
			}
		}
		close(bmutex)
	}()
	amutex <- true
	wg.Wait()

}

// 多线程应该注意：
//不能向已关闭的通道分发令牌---对输出长度长的增加分发限制，
//长竞争者如果是一轮竞争的前者需要保证短竞争者完成输出；
//如果竞争值只差1，那么谁结尾谁停发令牌即可。
