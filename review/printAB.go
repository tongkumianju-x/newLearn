package main

import (
	"fmt"
	"sync"
)

func main() {

	var a []int
	var b []byte

	for i := 0; i < 30; i++ {
		a = append(a, i+1)
	}

	for i := 0; i < 31; i++ {
		b = append(b, byte('A'+i))
	}

	printAB(a, b, 3)

}

func printAB(a []int, b []byte, size int) {
	var wg sync.WaitGroup

	atag := make(chan bool)
	btag := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(a); i += size {
			// 2.剩余部分全部输出,此时的i已经叠加了，直接用i比较长度即可,并且每次输出continue,并且输出前需等待1s，等待后竞争者的消费
			// 方法1：延迟等待方式
			//if len(a) > len(b) && i > len(b)-1 {
			//    for j := 0; j < size && j+i < len(a); j++ {
			//        time.Sleep(1 * time.Second)
			//        fmt.Print(a[i+j])
			//        continue
			//    }
			//    continue
			//}
			// 方法2: 计算出准确的i满足b完全发送完毕的情况，就无需延迟，但是需要在发送命令牌做限制;
			// 如果b是len=24,对应同位置是a[23]=24不能被纳入，a[24]=25也不应该被纳入(最后一次竞争)；那么i应该是26,
			// 如果b是len=25，那么a应该是a[24]=25,a[25]=26不能被纳入，a[26]=27
			if len(a) > len(b) && i > len(b)+len(b)%size {
				for j := 0; j < size && j+i < len(a); j++ {
					fmt.Print(a[i+j])
					continue
				}
				continue
			}
			<-atag
			for j := 0; j < size && j+i < len(a); j++ {
				fmt.Print(a[i+j])
			}
			// 1. 当a的数组等于b数组时不需要做特殊处理，因为a是先拿的

			// 2. 当a的数组大于b数组时，当a输出到b不能在输出的条件时，不再发送令牌
			// 属于方法2
			if len(a) > len(b) && i >= len(b) {
				continue
			}
			btag <- true
		}
		close(atag)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(b); i += size {
			// 2.剩余部分全部输出,此时的i已经叠加了，直接用i比较长度即可,并且每次输出continue
			if len(b) > len(a) && i > len(a)-1 {
				for j := 0; j < size && j+i < len(b); j++ {
					fmt.Print(string(b[i+j]))
					continue
				}
				continue
			}
			<-btag
			for j := 0; j < size && j+i < len(b); j++ {
				fmt.Print(string(b[i+j]))
			}
			// 1.当a的数组和b的数组长度一样时 b输出完应该保证b不再发送令牌
			if len(a) == len(b) && i+size >= len(a)-1+len(a)%size {
				continue
			}
			// 2.当b的数组大于a的数组时 需要保证在b发送至a的数组长度时不再发送令牌并且不再堵塞直接输出后半部分数据
			if len(b) > len(a) && i+size >= len(a)-1+len(a)%size {
				continue
			}
			// 3.当b的数组小于a的数组时 需要保证b最后一次不再发送令牌 因为b打印完后a进入超长输出不再进行阻塞
			// 方法1：
			// if len(b) < len(a) && i+size >= len(b)-1+len(b)%size {
			//    print("不再发送令牌")
			//    continue
			//}
			// 3.也可以设置代码不进入超长输出 那么需要提前设置 方法2

			atag <- true
		}
		close(btag)
	}()

	// 启动a线程
	atag <- true

	wg.Wait()
}
