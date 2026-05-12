package main

import (
	"fmt"
	"sync"
	"time"
)

type Philosopher struct {
	id        int
	leftFork  *sync.Mutex // 带上指针才能竞争
	rightFork *sync.Mutex
}

func (p *Philosopher) eat() {
	for {
		// 先拿左边的叉子
		p.leftFork.Lock()
		fmt.Printf("哲学家%d拿起了左边的叉子\n", p.id)

		// 模拟一些思考时间
		time.Sleep(time.Millisecond * 100)

		// 尝试拿右边的叉子
		p.rightFork.Lock()
		fmt.Printf("哲学家%d拿起了右边的叉子，开始就餐\n", p.id)

		// 就餐时间
		time.Sleep(time.Millisecond * 200)
		fmt.Printf("哲学家%d就餐完成\n", p.id)

		// 放下叉子
		p.rightFork.Unlock()
		p.leftFork.Unlock()

		// 思考时间
		time.Sleep(time.Millisecond * 300)
		break
	}
}

func main() {
	// 创建5把叉子（互斥锁）
	numPhilosophers := 5
	forks := make([]*sync.Mutex, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		forks[i] = &sync.Mutex{}
	}
	// 创建5个哲学家
	philosophers := make([]*Philosopher, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		philosophers[i] = &Philosopher{
			id:        i + 1,
			leftFork:  forks[i],                     // 左边的叉子
			rightFork: forks[(i+1)%numPhilosophers], // 右边的叉子（环形排列）
		}
		// 解锁方式1：混入左撇子
		//if i != numPhilosophers-1 {
		//    philosophers[i] = &Philosopher{
		//        id:        i + 1,
		//        leftFork:  forks[(i+1)%numPhilosophers], // 右边的叉子
		//        rightFork: forks[i],                     // 左边的叉子（环形排列）
		//    }
		//} else {
		//    philosophers[i] = &Philosopher{
		//        id:        i + 1,
		//        leftFork:  forks[i],                     // 左边的叉子
		//        rightFork: forks[(i+1)%numPhilosophers], // 右边的叉子（环形排列）
		//    }
		//}
		// 解锁方法2：奇偶数，一半左撇子，一半右撇子
		//
	}

	// 启动所有哲学家
	var wg sync.WaitGroup
	for i := 0; i < numPhilosophers; i++ {
		wg.Add(1)
		go func(p *Philosopher) {
			defer wg.Done()
			p.eat()
		}(philosophers[i])
	}

	// 运行一段时间后停止
	time.Sleep(time.Second * 5)
	fmt.Println("程序运行结束（可能已经发生死锁）")
}

// 死锁的必要条件：也是分析是否死锁的关键点 ：
// 1.互斥条件 ： 一个资源只能被一个线程获取
// 2.不可剥夺条件： 资源被获取后不能被强制剥夺
// 3.循环等待： 等待的资源形成循环
// 4.请求且保持 ： 等待下一个资源过程中不释放当前资源
