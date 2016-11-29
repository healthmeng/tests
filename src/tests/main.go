package main

import (
"fmt"
"os"
)

func prtUsage(){
	fmt.Println("Usege:")
}

func tryCreate(){
	fmt.Println("create:")
}

func listProj(){
	fmt.Println("list:")
}

func tryEdit(){
	fmt.Println("edit:")
}

func updateSrc(){
	fmt.Println("update:")
}

func main(){
	argc:=len(os.Args)
	if argc<2{
		prtUsage()
	}else{
		switch os.Args[1]{
		case "create":
			fallthrough
		case "-c"
			tryCreate()
		case "list":
			fallthrough
		case "-l"
			listProj()
		case "list":
			fallthrough
		case "-l"
			listProj()
		}
	}
}
