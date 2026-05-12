package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type DataProxyService struct{}

func (s *DataProxyService) GetLeaderBoard(uid float32, bid *float32) (ret int32, err error) {
	fmt.Printf("uid+ %f ", uid)
	fmt.Println()
	*bid = uid
	return 0, nil
}
func (s *DataProxyService) GetLeaderBoard2(uid int, bid *int) (ret int32, err error) {
	fmt.Printf("uid+ %d ", uid)
	fmt.Println()
	*bid = uid
	return 0, nil
}

func main() {
	s := &DataProxyService{}
	typ := reflect.TypeOf(s) //NumMethod 可以查看方法数量
	method, _ := typ.MethodByName("GetLeaderBoard")

	fmt.Println(method.Type.NumIn())
	argType := method.Type.In(1)
	arg := reflect.New(argType)
	// 创建一个新的变量
	newValue := reflect.New(argType)

	typ2 := reflect.TypeOf(float32(0.0))
	fmt.Println("typ Type:", typ2)
	if typ2.Kind() == reflect.Ptr {
		typ2 = typ2.Elem()
	}
	resp := reflect.New(typ2)

	if newValue.Elem().Kind() == reflect.Float32 {
		fmt.Println("newValue Type:", newValue)
		fmt.Println("oldValue:", newValue.Elem())
		data := []byte(`1`)
		_ = json.Unmarshal(data, newValue.Interface()) //另一种set方法
		fmt.Println("newValue:", newValue.Elem())
		method, _ = typ.MethodByName("GetLeaderBoard")
		_ = method.Func.Call([]reflect.Value{reflect.ValueOf(s), newValue.Elem(), resp})

		fmt.Println("123", resp.Elem())
		fmt.Println("123", arg.Kind())
	}
	fmt.Println("arg kind:", arg.Kind())

}
