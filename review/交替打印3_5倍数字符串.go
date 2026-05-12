package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type NumberPrint struct {
	Number              int
	nowPrintNum         int
	NumberChannel       chan struct{}
	ThreeChannel        chan struct{}
	FiveChannel         chan struct{}
	FiveAndThreeChannel chan struct{}
}

func initNumberPrint(n int) *NumberPrint {
	c := &NumberPrint{
		Number:              n,
		nowPrintNum:         1,
		NumberChannel:       make(chan struct{}, 1),
		ThreeChannel:        make(chan struct{}),
		FiveChannel:         make(chan struct{}),
		FiveAndThreeChannel: make(chan struct{}),
	}
	c.NumberChannel <- struct{}{}
	return c
}

func main() {
	var a []int
	for i := 1; i < 16; i++ {
		a = append(a, i)
	}
	threeAndFivePrint(a)
	fmt.Println()
	fmt.Println(strings.Repeat("-", 60))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// 读取用户输入
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		// 验证输入是否为数字
		n, _ := strconv.Atoi(input)

		c := initNumberPrint(n)

		c.threeAndFivePrints()
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("读取输入时出错: %v\n", err)
	}
	c := initNumberPrint(8)

	c.threeAndFivePrints()

}

// threeAndFivePrint 单线程
func threeAndFivePrint(a []int) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= len(a); i++ {
			if i%3 == 0 && i%5 != 0 {
				fmt.Print("fizz", ",")
			}
			if i%3 != 0 && i%5 == 0 {
				fmt.Print("buzz", ",")
			}
			if i%3 == 0 && i%5 == 0 {
				fmt.Print("fizzbuzz", ",")
			}
			if i%3 != 0 && i%5 != 0 {
				fmt.Print(i, ",")
			}
		}
	}()

	//wg.Add(1)
	//go func() {
	//    defer wg.Done()
	//    for i:=0;i<len(a);i++{
	//        if i%3!=0 && i%5==0{
	//            fmt.Print("buzz",",")
	//        }
	//    }
	//}()
	//
	//wg.Add(1)
	//go func() {
	//    defer wg.Done()
	//    for i:=0;i<len(a);i++{
	//        if i%3==0 && i%5==0{
	//            fmt.Print("fizzbuzz",",")
	//        }
	//    }
	//}()
	//
	//wg.Add(1)
	//go func() {
	//    defer wg.Done()
	//    for i:=0;i<len(a);i++{
	//        if i%3!=0 && i%5!=0{
	//            fmt.Print(i,",")
	//        }
	//    }
	//}()

	wg.Wait()
}

// threeAndFivePrints 多线程
func (c *NumberPrint) threeAndFivePrints() {
	var wg sync.WaitGroup

	// Number线程
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < c.Number-c.Number/3-c.Number/5+c.Number/15; i++ {
			<-c.NumberChannel
			fmt.Print(c.nowPrintNum, ",")
			c.nowPrintNum++
			if c.nowPrintNum%3 == 0 && c.nowPrintNum%5 == 0 && c.nowPrintNum <= c.Number {
				c.FiveAndThreeChannel <- struct{}{}
				//<-c.NumberChannel
			} else if c.nowPrintNum%3 != 0 && c.nowPrintNum%5 == 0 && c.nowPrintNum <= c.Number {
				c.FiveChannel <- struct{}{}
				//<-c.NumberChannel
			} else if c.nowPrintNum%3 == 0 && c.nowPrintNum%5 != 0 && c.nowPrintNum <= c.Number {
				c.ThreeChannel <- struct{}{}
				//<-c.NumberChannel
			} else if c.nowPrintNum%3 != 0 && c.nowPrintNum%5 != 0 && c.nowPrintNum <= c.Number {
				c.NumberChannel <- struct{}{}
			}
		}
		close(c.FiveChannel)
		close(c.FiveAndThreeChannel)
		close(c.ThreeChannel)
		close(c.NumberChannel)
	}()

	// 3倍数字线程
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < c.Number/3-c.Number/15; i++ {
			<-c.ThreeChannel
			fmt.Print("fizz", ",")
			c.nowPrintNum++
			if c.nowPrintNum%5 == 0 && c.nowPrintNum <= c.Number {
				c.FiveChannel <- struct{}{}
			} else if c.nowPrintNum <= c.Number {
				c.NumberChannel <- struct{}{}
			}
		}
	}()

	// 5倍数字线程
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < c.Number/5-c.Number/15; i++ {
			<-c.FiveChannel
			fmt.Print("buzz", ",")
			c.nowPrintNum++
			if c.nowPrintNum%3 == 0 && c.nowPrintNum <= c.Number {
				c.ThreeChannel <- struct{}{}
			} else if c.nowPrintNum <= c.Number {
				c.NumberChannel <- struct{}{}
			}
		}
	}()

	//3与5倍数字线程
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < c.Number/15; i++ {
			<-c.FiveAndThreeChannel
			fmt.Print("fizzbuzz", ",")
			c.nowPrintNum++
			if c.nowPrintNum <= c.Number {
				c.NumberChannel <- struct{}{}
			}
		}
	}()
	// c.NumberChannel <- struct{}{}

	wg.Wait()
}
