package main

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"
)

func main() {
	// 1. 普通 Pool 示例
	fmt.Println("========== 1. 普通 Pool ==========")
	normalPoolDemo()

	time.Sleep(300 * time.Millisecond)
	fmt.Println("\n========== 2. PoolWithFunc ==========")
	poolWithFuncDemo()

	time.Sleep(300 * time.Millisecond)
	fmt.Println("\n========== 3. PoolWithFuncGeneric ==========")
	poolWithFuncGenericDemo()
}

// 1. 普通 Pool：提交任意 func()
func normalPoolDemo() {
	// 池容量 3，空闲 2s 回收，带 panic 处理
	pool, err := ants.NewPool(3,
		ants.WithExpiryDuration(2*time.Second),
		ants.WithPanicHandler(func(i any) {
			fmt.Printf("panic 被捕获: %v\n", i)
		}),
	)
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	var wg sync.WaitGroup

	// 提交 5 个任务
	for i := 0; i < 5; i++ {
		i := i
		wg.Add(1)

		err := pool.Submit(func() {
			defer wg.Done()
			fmt.Printf("普通任务: %d\n", i)

			// 模拟一个 panic
			if i == 2 {
				panic("task 2 panic")
			}
			time.Sleep(100 * time.Millisecond)
		})

		if err != nil {
			fmt.Printf("提交任务失败: %v\n", err)
			wg.Done()
		}
	}

	wg.Wait()
	fmt.Printf("池容量: %d, 正在运行: %d, 空闲: %d\n", pool.Cap(), pool.Running(), pool.Free())

	// 动态调大池容量
	pool.Tune(5)
	fmt.Printf("调参后池容量: %d\n", pool.Cap())
}

// 2. PoolWithFunc：固定处理函数，参数 any
func poolWithFuncDemo() {
	// 固定处理函数
	process := func(args any) {
		id := args.(int)
		fmt.Printf("固定函数任务: %d\n", id)
		time.Sleep(100 * time.Millisecond)
	}

	pool, err := ants.NewPoolWithFunc(3, process,
		ants.WithExpiryDuration(2*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		i := i
		wg.Add(1)

		go func() {
			defer wg.Done()
			_ = pool.Invoke(i)
		}()
	}

	wg.Wait()
}

// 3. PoolWithFuncGeneric：泛型，类型安全
func poolWithFuncGenericDemo() {
	// 处理 int 类型
	process := func(id int) {
		fmt.Printf("泛型任务: %d\n", id)
		time.Sleep(100 * time.Millisecond)
	}

	pool, err := ants.NewPoolWithFuncGeneric[int](3, process,
		ants.WithExpiryDuration(2*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer pool.Release()

	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		i := i
		wg.Add(1)

		go func() {
			defer wg.Done()
			_ = pool.Invoke(i)
		}()
	}

	wg.Wait()
}
