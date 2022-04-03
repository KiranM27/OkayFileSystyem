package main

import (
	"fmt"
	"sync"
	//client "oks/Client"
)

var s sync.Map
func r(){
	s.Store("s","sd")
	
	temp := map[string]interface{}{}
	s.Range(func(k , v interface{}) bool {
		temp[fmt.Sprint(k)] = v
		return true
	})
	fmt.Println(temp)
	
	
}


