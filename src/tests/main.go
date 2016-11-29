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
	if len(os.Args)>2{
		prtUsage()
	}else{
		doList()
	}
}

func tryEdit(){
	fmt.Println("edit:")
}

func updateSrc(){
	fmt.Println("update:")
}

func delProj(){
    fmt.Println("del:")
}   

func main(){
	argc:=len(os.Args)
	if argc<2 || argc>3{
		prtUsage()
	}else{
		switch os.Args[1]{
		case "create":
			fallthrough
		case "-c":
			tryCreate()

		case "edit":
			fallthrough
		case "-e":
			tryEdit()

		case "list":
			fallthrough
		case "-l":
			listProj()

		case "update":
			fallthrough
		case "-u":
			updateSrc()

		case "del":
			fallthrough
		case "-d":
			delProj()

		default:
			prtUsage()
		}
	}
}
