package main

import (
    "fmt"
    "math"
)

func main() {

    a := "ADOBECODEBANC"
    t := "ABC"
    fmt.Print(piper(a, t))

}

func piper(s string, t string) string {
    tMap := make(map[byte]int)
    for i := 0; i < len(t); i++ {
        tMap[t[i]]++
    }
    tag := 0
    first := 0
    resultLen := math.MaxInt
    var result string
    for i := 0; i < len(s); i++ {
        v, ok := tMap[s[i]]
        if ok && v > 0 {
            tMap[s[i]]--
            tag++
        } else if ok && v == 0 {
            for ; s[first] != s[i]; first++ {
                _, ok2 := tMap[s[first]]
                if ok2 {
                    tMap[s[first]]++
                    tag--
                }
            }
            if s[first] == s[i] {
                first++
                for ; s[first] != s[i]; first++ {
                    _, ok2 := tMap[s[first]]
                    if ok2 {
                        break
                    }
                }
            }
        } else {
            continue
        }

        if tag == len(t) {
            if i-first+1 < resultLen {
                result = s[first : i+1]
                resultLen = i - first + 1
            }
        }

    }

    return result
}

//func min(x, y int) int {
//    if x > y {
//        return y
//    }
//    return x
//}
