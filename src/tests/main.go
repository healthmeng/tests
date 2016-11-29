package main

import (
"fmt"
"os"
)

func prtUsage(){
	fmt.Println("Tests is a tool for collecting your test codes and makes it easy for later review.")
	fmt.Println("Usege:")
	fmt.Println("\ttests command [args]")
	fmt.Println("commands include:")
	fmt.Println("\tcreate, -c dir|file: Create a test project.")
	fmt.Println("\tlist, -l: List all test projects and their infomations(IDs may be most useful.")
	fmt.Println("\tedit, -e proj_id: Edit an existing project.")
	fmt.Println("\tupdate, -u proj_id [dir|file]: Update codes of an existing project.If path is not refered, try to use current dir.")
	fmt.Println("\tdel, -d proj_id: Delete a project.")
}

func tryCreate(){
	fmt.Println("create:")
}

func listProj(){
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
	if argc<2 || argc>4{
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
