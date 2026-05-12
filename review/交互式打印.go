package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// 显示提示信息
		fmt.Println("\n=== 数字字符串生成器 ===")
		fmt.Println("请输入一个正整数 n (输入 'q' 退出程序):")

		// 读取用户输入
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// 检查是否退出
		if input == "q" || input == "quit" || input == "exit" {
			fmt.Println("程序退出，再见！")
			break
		}

		// 验证输入是否为数字
		n, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("错误：'%s' 不是有效的数字，请重新输入\n", input)
			continue
		}

		// 验证数字范围
		if n <= 0 {
			fmt.Printf("错误：请输入正整数，您输入的是 %d\n", n)
			continue
		}

		if n > 1000 {
			fmt.Printf("警告：数字 %d 较大，输出可能会很长，确定继续吗？(y/n): ", n)
			if !scanner.Scan() {
				break
			}
			confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if confirm != "y" && confirm != "yes" {
				fmt.Println("已取消")
				continue
			}
		}

		// 生成并打印字符串
		fmt.Printf("生成的字符串 (1-%d): ", n)
		printNumberStringFormatted(n)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取输入时出错: %v\n", err)
	}
}

// printNumberString 打印1到n的数字字符串
func printNumberString(n int) {
	for i := 1; i <= n; i++ {
		fmt.Print(i)
		if i < n {
			fmt.Print(" ") // 数字之间用空格分隔
		}
	}
	fmt.Println() // 换行
}

// 可选：另一种格式的打印函数
func printNumberStringFormatted(n int) {
	fmt.Printf("[")
	for i := 1; i <= n; i++ {
		fmt.Print(i)
		if i < n {
			fmt.Print(", ")
		}
	}
	fmt.Println("]")
}
