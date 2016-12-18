// process commondline arguments and invoke different actions

package main

import (
	"fmt"
	"os"
)

func prtUsage() {
	fmt.Println("Tests is a tool to collect your test codes and makes it easy for later review.")
	fmt.Println("Usage:")
	fmt.Println("\ttests command [args]")
	fmt.Println("commands include:")
	fmt.Println("\tcreate, -c dir|file: Create(upload) a test project.")
	fmt.Println("\tdel, -d proj_id: Delete a project.")
	fmt.Println("\tedit, -e proj_id: Edit an existing remote project's info.")
	fmt.Println("\tget, -g proj_id srcfile: Download a source file from project, currently support single regular file only.")
	fmt.Println("\tlist, -l | -l [id]: List all test projects and their infomations(IDs may be most useful) OR list project directory of certain id.")
	fmt.Println("\trun, -r proj_id [arg1 arg2 arg3...]: Run a proj, get commandline result. The args are only valid when project is a single source file.")
	fmt.Println("\tsearch, -s keyword1[, keyword2,keyword3...]: Search a project by keywords.")
	fmt.Println("\tupdate, -u proj_id [dir|file]: Update codes of an existing project.If path is not refered, try to use current dir.")
}

func tryCreate() {
	fmt.Println("create:")
	count := len(os.Args)
	if count > 3 {
		prtUsage()
	} else {
		path := os.Args[2]
		if finfo, err := os.Stat(path); err != nil {
			fmt.Println("Path error:", err.Error())
		} else {
			if id, err := doCreate(path, finfo.IsDir()); err != nil {
				fmt.Println("Create failed:", err)
			} else {
				fmt.Println("Create success,project id: ", id)
			}
		}
	}
}

func searchProj() {
}

func listProj() {
	nArgs:= len(os.Args)
	if nArgs > 3 {
		prtUsage()
	} else if nArgs == 3 {
		var id int64
		if _, err := fmt.Sscanf(os.Args[2], "%d", &id); err != nil {
			fmt.Println("Bad parameter:", os.Args[2])
		}else{
			doBrowse(id)
		}
	} else {
		doList()
	}
}

func tryEdit() {
	if len(os.Args) != 3 {
		prtUsage()
	} else {
		var id int64
		if _, err := fmt.Sscanf(os.Args[2], "%d", &id); err != nil {
			fmt.Println("Bad parameter:", os.Args[2])
		} else {
			doEdit(id)
		}
	}
}

func updateSrc() {
	fmt.Println("update:")
}

func delProj() {
	if len(os.Args) != 3 {
		prtUsage()
	} else {
		var id int64
		if _, err := fmt.Sscanf(os.Args[2], "%d", &id); err != nil {
			fmt.Println("Bad parameter:", os.Args[2])
		} else {
			doDel(id)
		}
	}
}

func getProj() {
	if len(os.Args)!=4 {
		prtUsage()
	}else{
		var id int64
		if _,err:=fmt.Sscanf(os.Args[2],"%d",&id);err!=nil{
			fmt.Println("Bad parameter:",os.Args[2])
			return
		}
		doGetFile(id,os.Args[3])
	}
}

func runProj() {
	nArg := len(os.Args)
	if nArg < 3 {
		prtUsage()
	} else {
		var id int64
		if _, err := fmt.Sscanf(os.Args[2], "%d", &id); err != nil {
			fmt.Println("Bad parameter:", os.Args[2])
			return
		}
		if nArg > 3 {
			doRun(id, os.Args[3:])
		} else {
			doRun(id, nil)
		}
	}
}

func main() {
	argc := len(os.Args)
	if argc < 2 {
		prtUsage()
	} else {
		switch os.Args[1] {
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

		case "get":
			fallthrough
		case "-g":
			getProj()

		default:
			prtUsage()
		}
	}
}
