package main

import (
	"fmt"
	"reflect"
)

type people struct {
	Name string `json:"name"`
	Age  int
}

func (s *people) SetName(name string) (err error) {
	s.Name = name
	return nil
}

// SetAge 设置年龄
func (s *people) SetAge(age int) (err error) {
	s.Age = age
	return nil
}

func (s *people) GetName() string {
	return s.Name
}

func (s *people) GetAge() int {
	return s.Age
}

func main() {

	type MyInt int
	var a MyInt = 10
	t := reflect.TypeOf(a)
	v := reflect.ValueOf(a)

	// t可调用的方法
	fmt.Println(t)        // 静态类型
	fmt.Println(t.Kind()) // 底层类型
	fmt.Println(t.Name()) // 类型名--MyInt

	b := people{}
	t1 := reflect.TypeOf(b)
	v1 := reflect.ValueOf(b)

	// t可调用的方法
	fmt.Println(t1)                      // 静态类型
	fmt.Println(t1.Kind())               // 底层类型
	fmt.Println(t1.Name())               // 类型名
	for i := 0; i < t1.NumField(); i++ { // Field来切分结构体的每个结构
		field := t1.FieldByIndex([]int{i}) //等同t1.Field(i)，FieldByIndex拥有层层解析能力
		fmt.Println(field)
		fmt.Println(field.Name) // 名称
		fmt.Println(field.Tag)  // json标记
		fieId, _ := t1.FieldByName(field.Name)
		fmt.Println(fieId)
		//StructField核心字段‌：
		//Name: 字段名
		//Type: 字段类型（reflect.Type）
		//Tag: 结构体标签（如 json:"name"），可通过 Get() 解析
		//Offset: 字段相对于结构体起始地址的字节偏移量
		//Anonymous: 是否为匿名字段
		//Index: 嵌套路径索引（用于多层嵌套）
	}
	b.SetName("alex")
	b.SetAge(18)
	t2 := reflect.TypeOf(&b)              //指针接收者方法(所有大写方法导出)  t2 := reflect.TypeOf(b) 是值接收者方法导出，小写开头的方法不导出
	for i := 0; i < t2.NumMethod(); i++ { // Method来切分结构体的每个方法
		methodId := t2.Method(i)
		fmt.Println(methodId.Name) // 名称
		fmt.Println(methodId.Type) // 含接收者、参数、返回值
		// Type延展
		fmt.Println(methodId.Type.NumIn())
		fmt.Println(methodId.Type.In(0))

		fmt.Println(methodId.Index)
		fmt.Println(methodId.Func) //是reflect.Value类型
		MethodId, _ := t2.MethodByName(methodId.Name)
		fmt.Println(MethodId)
		//    Method核心字段
		//    Name string // 方法名（仅导出方法可见）
		//    PkgPath string // 包路径（用于非导出方法的唯一标识）
		//    Type reflect.Type // 方法的类型（含接收者、参数、返回值）
		//    Func reflect.Value // 方法的函数值，可直接调用
		//    Index int // 方法在类型方法集中的索引
	}

	// v可调用的方法
	fmt.Println(v)               //查值 v.Int()/ v.String()/ v.Bool()/ v.Float()
	fmt.Println(v.Type())        //依旧是静态类型
	fmt.Println(v.Type().Kind()) //底层类型
	fmt.Println(v.Kind())        //底层类型
	fmt.Println(v.Type().Name()) //类型名 没有v.Name()

	v2 := reflect.ValueOf(&a).Elem() //v2 = reflect.ValueOf(a)不能被setInt
	fmt.Println(v2)
	fmt.Println(v2.CanSet())
	fmt.Println(v2.Type()) //依旧是静态类型
	fmt.Println(v2.Kind()) //底层类型
	v2.SetInt(20)
	fmt.Println(a)

	s := []int{1, 2, 3}
	v3 := reflect.ValueOf(s)
	fmt.Println(v3.Len())
	v3.Index(0).SetInt(100)
	fmt.Println(v3)

	m := map[string]int{"a": 1}
	v4 := reflect.ValueOf(m)
	v4.SetMapIndex(reflect.ValueOf("a"), reflect.ValueOf(100))
	//v4.MapIndex(reflect.ValueOf("a"))
	fmt.Println(v4)
	fmt.Println(v4.MapIndex(reflect.ValueOf("a")))

	u := people{Name: "Tom"}
	//ux := reflect.ValueOf(u).Elem() //报错
	//fmt.Println(ux)
	//nameFieldx := ux.FieldByName("Name")
	//nameFieldx.SetString("Jerry")
	//fmt.Println(nameFieldx)
	u1 := reflect.ValueOf(&u).Elem()
	fmt.Println(u1)
	nameField := u1.FieldByName("Name")
	nameField.SetString("Jerry")
	fmt.Println(nameField)

	fmt.Println(v1)
	fmt.Println(v1.Type()) //依旧是静态类型
	fmt.Println(v1.Kind()) //底层类型

	//new--->指针
	u2 := reflect.New(t)
	u2.Elem().SetInt(10)
	fmt.Println(u2.Elem())

}
