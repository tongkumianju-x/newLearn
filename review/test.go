package main

import "fmt"

func main() {
	var uidList []string
	fmt.Println(uidList)
	if len(uidList) == 0 {
		fmt.Println(2)
	}
	if uidList == nil {
		fmt.Println(1)
	}
}
