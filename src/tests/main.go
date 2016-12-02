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
	fmt.Println("\tedit, -e proj_id: Edit an existing project's info.")
	fmt.Println("\tupdate, -u proj_id [dir|file]: Update codes of an existing project.If path is not refered, try to use current dir.")
	fmt.Println("\tdel, -d proj_id: Delete a project.")
	fmt.Println("\tsearch, -s keyword1[, keyword2,keyword3...]: Search a project by keywords.")
	fmt.Println("\trun, -r proj_id: Run a proj, get commandline result.")
}

func tryCreate(){
	fmt.Println("create:")
	count:=len(os.Args)
	if(count>3){
		prtUsage()
	}else{
		path:=os.Args[2]
		if finfo,err:=os.Stat(path);err!=nil{
			fmt.Println("Path error:",err.Error())
		}else{
			if id,err:=doCreate(path,finfo.IsDir());err!=nil{
				fmt.Println("Create failed:",err)
			}else{
				fmt.Println("Create success,project id: ",id)
			}
		}
	}
}

func searchProj(){
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


func runProj(){
    fmt.Println("run:")
}


func main(){
	argc:=len(os.Args)
	if argc<2 {
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

		case "search":
			fallthrough
		case "-s":
			searchProj()

		case "run":
			fallthrough
		case "-r":
			runProj()

		default:
			prtUsage()
		}
	}
}
