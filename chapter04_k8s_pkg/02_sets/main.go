package main


import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
)

func main(){
	map1 := map[string]int{"bbb":2,"aaa":1,"ccc":3}
	map2 := map[string]int{"ddd":2,"eee":3,"ccc":1}
	// 构建Map
	newmap1 := sets.StringKeySet(map1)
	newmap2 := sets.StringKeySet(map2)
	fmt.Println(newmap1.List(),newmap2.List()) // 打印排序后的切片
	fmt.Println(newmap1.HasAny(newmap2.List()...)) //3个点用于把数组打散为单个元素
}
