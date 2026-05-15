package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

func main() {
	testList := rankListGet(100)
	var merageKList func(int, int) []int
	merageKList = func(left, right int) []int {
		if left == right {
			return testList[left]
		}
		mid := (left + right) / 2
		leftList := merageKList(left, mid)
		rightList := merageKList(mid+1, right)
		return merageList(leftList, rightList)
	}
	result := merageKList(0, len(testList)-1)
	fmt.Print(result)
}

func rankListGet(num int) [][]int {
	var nums [][]int
	rands := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 1; i <= num; i++ {
		var numj []int
		for j := 0; j < 100; j++ {
			numj = append(numj, rands.Intn(100))
		}
		sort.Ints(numj)
		nums = append(nums, numj)
	}
	return nums
}

//func merageKList(left [][]int,right [][]int) []int {
//}

func merageList(a []int, b []int) []int {

	var c []int
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] > b[j] {
			c = append(c, b[j])
			j++
		} else {
			c = append(c, a[i])
			i++
		}
	}
	if i < len(a) {
		c = append(c, a[i:]...)
	}
	if j < len(b) {
		c = append(c, b[j:]...)
	}
	return c
}
