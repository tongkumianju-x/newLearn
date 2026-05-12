package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func main() {
	var s string
	s = "barfoothefoobarman"
	var s2 []string
	s2 = append(s2, "foo")
	s2 = append(s2, "bar")
	x := 'a' - 'b'
	fmt.Println(reflect.TypeOf(x))
	fmt.Println(findSubstring(s, s2))
	TestStrings()
	TestStrConv()
	fmt.Println(isPalindrome("ab_a"))

}

func TestStrings() {
	// 判断类 ：含有包含 Contains
	fmt.Println(strings.Contains("abc", "a"))
	fmt.Println(strings.ContainsAny("hello", "le"))
	fmt.Println(strings.ContainsFunc("abc", func(r rune) bool {
		return r == 'd'
	}))
	fmt.Println(strings.ContainsRune("abc", 'a'))

	// 判断类：前后缀
	fmt.Println(strings.HasPrefix("alex", "a"))
	fmt.Println(strings.HasSuffix("alex", "ex"))

	// 判断类：索引 any byte
	fmt.Println(strings.Index("alexa", "l"))
	fmt.Println(strings.LastIndex("alexa", "a"))

	// 计数
	fmt.Println(strings.Count("aaaa", "aa"))

	// 切分 split splitN splitAfter 返回字符串数组
	fmt.Println(strings.Split("a,b,c", ","))
	// Fields 去除空格换行制表符号（包含中间） FieldsFunc特制广义分隔符
	fmt.Println(strings.Fields("   a,b,c  ")) //返回字符串数组
	fmt.Println(strings.FieldsFunc("1a2b3c", func(r rune) bool {
		return r >= 'a' && r <= 'z'
	}))

	// 拼接和替换
	// Join replace replaceAll

	// 重复 repeat

	// 大小写 ToLower ToUpper Title ToTitle

	// 修剪裁剪(首尾) TrimPrefix TrimSuffix  TrimLeft TrimRight Trim  TrimSpace
	fmt.Println(strings.TrimSpace("   a,b,c   ")) //返回字符串

	// Rune(1.21没有了) Map(特殊func)
	fmt.Println(strings.Map(func(r rune) rune {
		return r + 1
	}, "abc"))

}

func TestStrConv() {

	fmt.Printf("%T", strconv.Itoa(123))
	fmt.Println()
	// strconv
	n, _ := strconv.Atoi("123")
	fmt.Printf("%T", n)
	fmt.Println()

	fmt.Println(strconv.FormatFloat(1.64, 'f', 2, 64))
	fmt.Println(strconv.FormatFloat(1.64, 'e', 2, 64))
	fmt.Println(strconv.FormatFloat(1.64, 'g', 2, 64))

	n1, _ := strconv.ParseFloat("1.643", 10)
	fmt.Printf("%v ,%T", n1, n1)
	fmt.Println()

	fmt.Println(strconv.FormatInt(13, 10))
	n2, _ := strconv.ParseInt("1643", 10, 10) //转64位但是会按照bitsize位做检查
	fmt.Printf("%v ,%T", n2, n2)
	fmt.Println()
	fmt.Println(2 << 8)
}

func isPalindrome(s string) bool {
	result := strings.FieldsFunc(s, func(r rune) bool {
		return (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	fmt.Println(result)
	result2 := strings.Join(result, "")
	fmt.Println(result2)
	result3 := strings.ToLower(result2)
	fmt.Println(result3)

	if len(result3) == 0 || len(result3) == 1 {
		return true
	}
	if len(result3)%2 == 0 {
		return isPalindromes(result3, len(result3)/2-1, len(result3)/2)
	} else {
		return isPalindromes(result3, len(result3)/2, len(result3)/2)
	}

}

func isPalindromes(s string, i, j int) bool {
	for ; i >= 0 && j < len(s); i++ {
		if s[i] != s[j] {
			return false
		}
		j++
	}
	return true
}

func findSubstring(s string, words []string) []int {
	n := len(s)
	m := len(words)
	m1 := 0

	if m > 0 {
		m1 = len(words[0])
	}

	maxLength := m1 * m
	if n < maxLength {
		return []int{}
	}

	result := []int{}

	left := 0
	for left < n-maxLength {
		newpipeiwordMap := make(map[string]int)
		for _, j := range words {
			newpipeiwordMap[j]++
		}
		for i := left; i < left+maxLength; i += m1 {
			if newpipeiwordMap[s[i:i+m1]] > 0 {
				newpipeiwordMap[s[i:i+m1]]--
			} else {
				break
			}
			if i == left+maxLength-m1 {
				result = append(result, left)
			}
		}
		left++
	}

	return result
}
