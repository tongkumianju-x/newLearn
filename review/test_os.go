package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	w := bufio.NewWriter(os.Stdout)
	w.WriteString("hello world")
	w.Flush()
	fmt.Println()

	r := bufio.NewScanner(os.Stdin)
	r.Split(bufio.ScanWords) //注册分割方法 默认按行 lines/runes/bytes
	for r.Scan() {
		if r.Text() == "exit" {
			break
		}
		fmt.Println(r.Text())
	}
	//需要创建文件夹 0755指权限
	_ = os.Mkdir("os", 0755)

	f1, _ := os.Create("./os/open.txt") // 相对地址./xx/xx 绝对地址 /xx/xxx/xxx  当前地址 xxx.txt
	defer f1.Close()
	w = bufio.NewWriter(f1)
	w.WriteString("hello world")
	w.Flush()
	fmt.Println()

	f2, _ := os.Open("./os/open.txt")
	r = bufio.NewScanner(f2)
	for r.Scan() {
		fmt.Println(r.Text())
	}
	defer f2.Close()

}
