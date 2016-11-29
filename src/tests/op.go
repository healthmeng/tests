package main

import (
"fmt"
)

var remotesvr string ="123.206.55.31"

type PROJINFO struct{
	id int
	title string
	atime string
	descr string
	conclude string
	path string
}



func doList(){
	fmt.Println("Do list")
}

func doCreate(path string){
	fmt.Println(path)
}
