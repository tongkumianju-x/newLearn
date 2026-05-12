package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	fmt.Println(runtime.GOMAXPROCS(5))
	fmt.Println(runtime.NumCPU())
	fmt.Println(runtime.NumGoroutine())
	fmt.Println(runtime.GOOS)
	fmt.Println(runtime.GOARCH)

	//testPG()
	//testPG2()
	//testPG3()
	//testPG4()
	DemoGosched()
	runtime.GC() // 并发标记清除+混合写屏障+低STW--三色（黑灰白）--
	// 步骤：1。标记准备：开启写屏障，STW开始，标记根对象为灰色
	// 2。并发标记 灰-->黑 白--->灰
	// 3。二次扫描根 无灰即结束 关闭写屏障
	// 4。清除

}
func DemoGosched() {
	done := false

	runtime.GOMAXPROCS(1)
	// 启动一个后台协程
	go func() {
		fmt.Println("  [后台协程] 开始运行...")
		time.Sleep(100 * time.Millisecond)
		done = true
		fmt.Println("  [后台协程] 运行结束")
	}()
	// 启动一个后台协程
	go func() {
		fmt.Println("  [后台协程2] 开始运行...")
		time.Sleep(100 * time.Millisecond)
		done = true
		fmt.Println("  [后台协程2] 运行结束")
	}()

	// 模拟一个耗时循环
	for i := 0; i < 5; i++ {
		if done {
			break
		}
		fmt.Printf("  [主协程] 正在忙碌 %d...\n", i)
		// 如果不加 Gosched，在某些极端情况下，主协程可能一直霸占 CPU
		// 导致后台协程没机会执行（虽然 Go 调度器有抢占式机制，但 Gosched 是显式礼让）
		// runtime.Gosched()
	}
	// select {} //永久阻塞
	// time.Sleep(time.Second)
	// if err = s.Serve(); err != nil {
	//		logrus.Errorf("%v", err)
	//	}是典型的阻塞

}

//func testPG() {
//    // 设置P的数量（逻辑CPU核心数）
//    runtime.GOMAXPROCS(2)
//    var wg sync.WaitGroup
//
//    // 启动10个G
//    for i := 0; i < 10; i++ {
//        wg.Add(1)
//        go func(n int) {
//            defer wg.Done()
//            // 获取goroutine ID（需要hack方式）
//            var buf [64]byte
//            x := runtime.Stack(buf[:], false)
//            idField := strings.Fields(strings.TrimPrefix(string(buf[:x]), "goroutine "))[0]
//            id, _ := strconv.Atoi(idField)
//
//            fmt.Printf("Goroutine %d 执行中 | GID: %d\n",
//                n, id)
//        }(i)
//    }
//
//    wg.Wait()
//}
//func testPG2() {
//    // 锁定当前goroutine到当前线程
//    runtime.LockOSThread()
//    defer runtime.UnlockOSThread()
//
//    fmt.Println("g2 开始，准备阻塞（已锁定到当前线程）")
//
//    // 获取goroutine ID
//    var buf [64]byte
//    x := runtime.Stack(buf[:], false)
//    idField := strings.Fields(strings.TrimPrefix(string(buf[:x]), "goroutine "))[0]
//    gid, _ := strconv.Atoi(idField)
//
//    procs := runtime.GOMAXPROCS(0)
//    numGoroutine := runtime.NumGoroutine()
//
//    fmt.Printf("阻塞前 → GID:%d | P数量:%d | 总G数量:%d | 线程锁定:是\n",
//        gid, procs, numGoroutine)
//
//    // 阻塞！这里会触发调度
//    time.Sleep(2 * time.Second)
//
//    // 唤醒后
//    fmt.Println("g2 唤醒")
//
//    // 再次获取运行时信息
//    x = runtime.Stack(buf[:], false)
//    idField = strings.Fields(strings.TrimPrefix(string(buf[:x]), "goroutine "))[0]
//    gid, _ = strconv.Atoi(idField)
//
//    procs = runtime.GOMAXPROCS(0)
//    numGoroutine = runtime.NumGoroutine()
//
//    fmt.Printf("阻塞后 → GID:%d | P数量:%d | 总G数量:%d | 线程锁定:是\n",
//        gid, procs, numGoroutine)
//}
//// 耗时任务，让P本地队列占满
//func task(id int) {
//    for i := 0; i < 1000000000; i++ {
//    }
//    fmt.Printf("任务 %d 完成\n", id)
//}
//func testPG3() {
//    runtime.GOMAXPROCS(2) // 2个P
//    var wg sync.WaitGroup
//
//    // 启动大量G
//    for i := 0; i < 8; i++ {
//        wg.Add(1)
//        go func(n int) {
//            defer wg.Done()
//            task(n)
//        }(i)
//    }
//
//    // 等待所有G执行完
//    time.Sleep(3 * time.Second)
//    wg.Wait()
//}
//func testPG4() {
//    // 只开 1 个 P，这样最容易看出调度行为
//    runtime.GOMAXPROCS(1)
//
//    // 启动一个死循环 G 占用 P
//    go func() {
//        for {
//        }
//    }()
//
//    // 让主goroutine先让出，让死循环G先跑
//    runtime.Gosched()
//
//    // 再运行 testPG2，它会阻塞并展示调度行为
//    testPG2()
//}
