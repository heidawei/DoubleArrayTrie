package main

import (
	"fmt"

	"github.com/heidawei/DoubleArrayTrie/darts"
)

func main() {
	words := []string{"一举", "一举一动", "一举成名", "一举成名天下知", "万能", "万能胶"}
	dat := darts.NewDoubleArrayTrie()
	fmt.Println("是否错误: ", dat.Build(words))
	//dat.Dump()
	index := dat.ExactMatchSearch("万能")
	if index >= 0 {
		fmt.Println(words[index])
	}
	index = dat.ExactMatchSearch("哈哈")
	if index >= 0 {
		fmt.Println(words[index])
	}
	fmt.Println("size ", dat.GetSize())
	list := dat.CommonPrefixSearch("一举成名天下知")
	for _, index := range list {
		fmt.Println(words[index])
	}
}
