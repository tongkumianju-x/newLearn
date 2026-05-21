package main

import "fmt"

// ListNode 单链表节点
type ListNode struct {
    Val  int
    Next *ListNode
}

// reorderList 重排链表：L0 -> L1 -> ... -> Ln  =>  L0 -> Ln -> L1 -> Ln-1 -> ...
// 思路三步走：
//  1. 快慢指针找中点，切成前后两段
//  2. 反转后半段
//  3. 交替合并前半段与反转后的后半段
//
// 时间 O(n)，空间 O(1)
func reorderList(head *ListNode) {
    // 边界：空链表或只有 1、2 个节点不需要重排
    if head == nil || head.Next == nil || head.Next.Next == nil {
        return
    }

    // 1) 快慢指针找中点
    //    奇数节点：slow 落在正中；偶数节点：slow 落在前半段最后一个
    slow, fast := head, head
    for fast.Next != nil && fast.Next.Next != nil {
        slow = slow.Next
        fast = fast.Next.Next
    }

    // 2) 切断并反转后半段
    second := slow.Next
    slow.Next = nil
    second = reverseList(second)

    // 3) 交替合并：first 段长度 >= second 段长度
    first := head
    for second != nil {
        fNext := first.Next
        sNext := second.Next
        first.Next = second
        second.Next = fNext
        first = fNext
        second = sNext
    }
}

// reverseList 反转单链表
func reverseList(head *ListNode) *ListNode {
    var prev *ListNode
    cur := head
    for cur != nil {
        next := cur.Next
        cur.Next = prev
        prev = cur
        cur = next
    }
    return prev
}

// buildList 用切片快速构造链表，方便写测试用例
func buildList(vals []int) *ListNode {
    dummy := &ListNode{}
    tail := dummy
    for _, v := range vals {
        tail.Next = &ListNode{Val: v}
        tail = tail.Next
    }
    return dummy.Next
}

// listToSlice 把链表转回切片，方便打印 / 比对
func listToSlice(head *ListNode) []int {
    res := []int{}
    for cur := head; cur != nil; cur = cur.Next {
        res = append(res, cur.Val)
    }
    return res
}

// equalSlice 比较两个 int 切片是否相等
func equalSlice(a, b []int) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func main() {
    // 3 组测试用例：奇数长度、偶数长度、边界（单节点）
    cases := []struct {
        name   string
        input  []int
        expect []int
    }{
        {
            name:   "奇数长度 5：1->2->3->4->5",
            input:  []int{1, 2, 3, 4, 5},
            expect: []int{1, 5, 2, 4, 3},
        },
        {
            name:   "偶数长度 6：1->2->3->4->5->6",
            input:  []int{1, 2, 3, 4, 5, 6},
            expect: []int{1, 6, 2, 5, 3, 4},
        },
        {
            name:   "边界：单节点",
            input:  []int{42},
            expect: []int{42},
        },
    }

    for _, c := range cases {
        head := buildList(c.input)
        reorderList(head)
        got := listToSlice(head)
        ok := equalSlice(got, c.expect)
        fmt.Printf("[%s]\n  input  = %v\n  expect = %v\n  got    = %v\n  pass   = %v\n\n",
            c.name, c.input, c.expect, got, ok)
    }
}
