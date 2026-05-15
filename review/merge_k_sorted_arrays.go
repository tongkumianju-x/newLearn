package main

import (
	"container/heap"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// ============================================================
// 一、最小堆方式合并 K 个有序数组
// ============================================================

// heapNode 表示堆中的一个元素，记录值、来源数组下标、该数组中元素下标
type heapNode struct {
	val     int // 当前元素的值
	arrIdx  int // 该元素所属的数组在 arrays 中的下标
	elemIdx int // 该元素在所属数组中的下标
}

// minHeap 实现 heap.Interface 接口，按 val 升序
type minHeap []heapNode

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].val < h[j].val }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(heapNode)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// MergeKArraysByHeap 使用最小堆合并 K 个有序数组
// 时间复杂度: O(N * logK)，其中 N 为所有元素总数，K 为数组个数
// 空间复杂度: O(K)，堆中最多保存 K 个元素
func MergeKArraysByHeap(arrays [][]int) []int {
	h := &minHeap{}
	heap.Init(h)

	// 总元素个数，用于预分配结果切片容量
	total := 0
	// 1. 把每个非空数组的第一个元素放入堆中
	for i, arr := range arrays {
		total += len(arr)
		if len(arr) > 0 {
			heap.Push(h, heapNode{val: arr[0], arrIdx: i, elemIdx: 0})
		}
	}

	result := make([]int, 0, total)
	// 2. 不断弹出堆顶，并将该元素所在数组的下一个元素入堆
	for h.Len() > 0 {
		top := heap.Pop(h).(heapNode)
		result = append(result, top.val)

		nextIdx := top.elemIdx + 1
		if nextIdx < len(arrays[top.arrIdx]) {
			heap.Push(h, heapNode{
				val:     arrays[top.arrIdx][nextIdx],
				arrIdx:  top.arrIdx,
				elemIdx: nextIdx,
			})
		}
	}
	return result
}

// ============================================================
// 二、分治方式合并 K 个有序数组
// ============================================================

// mergeTwoSortedArrays 合并两个有序数组（升序）
func mergeTwoSortedArrays(a, b []int) []int {
	merged := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			merged = append(merged, a[i])
			i++
		} else {
			merged = append(merged, b[j])
			j++
		}
	}
	if i < len(a) {
		merged = append(merged, a[i:]...)
	}
	if j < len(b) {
		merged = append(merged, b[j:]...)
	}
	return merged
}

// MergeKArraysByDivide 使用分治法合并 K 个有序数组
// 思路: 两两合并，类似归并排序的合并过程
// 时间复杂度: O(N * logK)
// 空间复杂度: O(N * logK)，每层递归都需要新建合并后的数组
func MergeKArraysByDivide(arrays [][]int) []int {
	if len(arrays) == 0 {
		return []int{}
	}
	return divideAndMerge(arrays, 0, len(arrays)-1)
}

// divideAndMerge 递归地把 arrays[left..right] 区间内的数组两两合并
func divideAndMerge(arrays [][]int, left, right int) []int {
	if left == right {
		// 注意：返回拷贝，避免外部修改影响到原数组
		res := make([]int, len(arrays[left]))
		copy(res, arrays[left])
		return res
	}
	mid := left + (right-left)/2
	l := divideAndMerge(arrays, left, mid)
	r := divideAndMerge(arrays, mid+1, right)
	return mergeTwoSortedArrays(l, r)
}

// ============================================================
// 三、测试数据生成与耗时比较
// ============================================================

// generateSortedArrays 随机生成 k 个有序数组，每个数组长度在 [minLen, maxLen] 之间
// 元素取值范围 [0, valueRange)
func generateSortedArrays(k, minLen, maxLen, valueRange int, seed int64) [][]int {
	rng := rand.New(rand.NewSource(seed))
	arrays := make([][]int, k)
	for i := 0; i < k; i++ {
		length := minLen + rng.Intn(maxLen-minLen+1)
		arr := make([]int, length)
		for j := 0; j < length; j++ {
			arr[j] = rng.Intn(valueRange)
		}
		sort.Ints(arr) // 保证每个子数组有序
		arrays[i] = arr
	}
	return arrays
}

// isSorted 校验结果是否升序
func isSorted(arr []int) bool {
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[i-1] {
			return false
		}
	}
	return true
}

// runBenchmark 针对一组参数运行两种算法并打印耗时
func runBenchmark(k, minLen, maxLen, valueRange int, seed int64) {
	arrays := generateSortedArrays(k, minLen, maxLen, valueRange, seed)

	// 统计总元素数
	total := 0
	for _, a := range arrays {
		total += len(a)
	}
	fmt.Printf("\n========== 测试用例: K=%d, 总元素数=%d ==========\n", k, total)

	// 1. 最小堆方式
	start := time.Now()
	resHeap := MergeKArraysByHeap(arrays)
	heapCost := time.Since(start)

	// 2. 分治方式
	start = time.Now()
	resDivide := MergeKArraysByDivide(arrays)
	divideCost := time.Since(start)

	// 校验正确性
	heapOK := isSorted(resHeap) && len(resHeap) == total
	divideOK := isSorted(resDivide) && len(resDivide) == total
	bothEqual := len(resHeap) == len(resDivide)
	if bothEqual {
		for i := range resHeap {
			if resHeap[i] != resDivide[i] {
				bothEqual = false
				break
			}
		}
	}

	fmt.Printf("最小堆算法耗时: %v  (有序性: %v)\n", heapCost, heapOK)
	fmt.Printf("分治算法耗时:  %v  (有序性: %v)\n", divideCost, divideOK)
	fmt.Printf("两种算法结果一致: %v\n", bothEqual)

	if heapCost < divideCost {
		fmt.Printf(">>> 最小堆较快，快 %.2fx\n", float64(divideCost)/float64(heapCost))
	} else {
		fmt.Printf(">>> 分治较快，快 %.2fx\n", float64(heapCost)/float64(divideCost))
	}
}

func main() {
	// 设计多组测试数据，对比两种算法在不同规模下的表现
	cases := []struct {
		k, minLen, maxLen, valueRange int
		seed                          int64
	}{
		{k: 10, minLen: 50, maxLen: 100, valueRange: 1000, seed: 1},
		{k: 50, minLen: 100, maxLen: 200, valueRange: 10000, seed: 2},
		{k: 100, minLen: 500, maxLen: 1000, valueRange: 100000, seed: 3},
		{k: 500, minLen: 500, maxLen: 1000, valueRange: 1000000, seed: 4},
		{k: 1000, minLen: 1000, maxLen: 2000, valueRange: 10000000, seed: 5},
	}

	for _, c := range cases {
		runBenchmark(c.k, c.minLen, c.maxLen, c.valueRange, c.seed)
	}
}
