package main

import "fmt"

type LeftRightTree struct {
	Val   int
	Left  *LeftRightTree
	Right *LeftRightTree
}

func BFS(root *LeftRightTree) []int {
	if root == nil {
		return []int{}
	}
	result := make([]int, 0)

	cunchu := []*LeftRightTree{root}
	for len(cunchu) > 0 {
		result = append(result, cunchu[0].Val)

		if cunchu[0].Left != nil {
			cunchu = append(cunchu, cunchu[0].Left)
		}
		if cunchu[0].Right != nil {
			cunchu = append(cunchu, cunchu[0].Right)
		}

		cunchu = cunchu[1:]

	}
	return result
}

func BFS2(root *LeftRightTree) [][]int {
	if root == nil {
		return [][]int{}
	}
	maxResult := make([][]int, 0)

	cunchu := []*LeftRightTree{root}
	for i := 0; len(cunchu) > 0; i++ {

		result := make([]int, 0)
		cunchu2 := []*LeftRightTree{}

		for j := 0; j < len(cunchu); j++ {
			result = append(result, cunchu[j].Val)

			if cunchu[j].Left != nil {
				cunchu2 = append(cunchu2, cunchu[j].Left)
			}
			if cunchu[j].Right != nil {
				cunchu2 = append(cunchu2, cunchu[j].Right)
			}
		}
		cunchu = cunchu2
		maxResult = append(maxResult, result)

	}
	return maxResult
}

// Pre 前序遍历
func Pre(root *LeftRightTree) []int {
	result := make([]int, 0)

	var pre func(root *LeftRightTree)
	pre = func(root *LeftRightTree) {
		if root == nil {
			return
		}
		// 中序，后序更换位置即可
		result = append(result, root.Val)
		pre(root.Left)
		pre(root.Right)
	}
	pre(root)
	return result

}

// DFS 非递归版本
func DFS(root *LeftRightTree) []int {
	if root == nil {
		return []int{}
	}
	result := make([]int, 0)

	cunchu := []*LeftRightTree{root}
	for len(cunchu) > 0 {
		node := cunchu[len(cunchu)-1]
		result = append(result, cunchu[len(cunchu)-1].Val)
		cunchu = cunchu[:len(cunchu)-1]

		if node.Right != nil {
			cunchu = append(cunchu, node.Right)
		}
		if node.Left != nil {
			cunchu = append(cunchu, node.Left)
		}

		//cunchu = cunchu[1:]

	}
	return result
}

func main() {

	root := &LeftRightTree{Val: 1}
	root.Left = &LeftRightTree{Val: 2}
	root.Right = &LeftRightTree{Val: 3}
	root.Left.Left = &LeftRightTree{Val: 4}
	root.Left.Right = &LeftRightTree{Val: 5}
	root.Right.Left = &LeftRightTree{Val: 3}
	root.Right.Right = &LeftRightTree{Val: 1}

	fmt.Printf("广度优先遍历:%v", BFS(root))

	fmt.Printf("广度优先遍历2:%v", BFS2(root))

	fmt.Printf("前序遍历：%v", Pre(root))

	fmt.Printf("DFS遍历：%v", DFS(root))

}
