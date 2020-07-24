// process commondline arguments and invoke different actions

package main

import (
	"fmt"
	"os"
)

func prtUsage() {
	fmt.Println("\n\ttests is a tool to collect your testing codes and make it easy for later review.\n")
	fmt.Println("Usage:")
	fmt.Println("\ttests command [args]\n")
	fmt.Println("Commands include:")
	fmt.Println("\tcreate, -c dir|file\n\t\tCreate(upload) a test project.\n")
	fmt.Println("\tdel, -d proj_id\n\t\tDelete a project.\n")
	fmt.Println("\tedit, -e proj_id:\n\t\tEdit an existing remote project's info.\n")
	fmt.Println("\tget, -g proj_id | -l proj_id srcfile(including relative directory)\n\t\tDownload the whole project to current directory OR dump a single source file in a project to stdout.\n")
	fmt.Println("\tlist, -l | -l id\n\t\tList all test projects and their infomations(IDs may be most useful) OR list project directory structures of a certain project by id.\n")
	fmt.Println("\trun, -r proj_id [arg1 arg2 arg3...]\n\t\tRun a proj, get commandline result. The args are valid only when project contains a single source file.\n")
	fmt.Println("\tsearch, -s keyword1[ keyword2 keyword3...]\n\t\tSearch a project by case-insensitive keywords.\n")
	fmt.Println("\tSearch, -S keyword1[ keyword2 keyword3...]\n\t\tSearch a project by case-sensitive keywords.\n")
	fmt.Println("\tupdate, -u proj_id proj_file(including relative directory) localfile\n\t\tUpdate a source file of an existed project.\n")
}

func tryCreate() {
	fmt.Println("create:")
	count := len(os.Args)
	if count !=3{
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
	nArgs := len(os.Args)
	if nArgs < 3 {
		prtUsage()
	} else {
		doSearch(os.Args[2:],false)
	}
}

func searchCSProj() {
	nArgs := len(os.Args)
	if nArgs < 3 {
		prtUsage()
	} else {
		doSearch(os.Args[2:],true)
	}
}

func listProj() {
	nArgs := len(os.Args)
	if nArgs > 3 {
		prtUsage()
	} else if nArgs == 3 {
		var id int64
		if _, err := fmt.Sscanf(os.Args[2], "%d", &id); err != nil {
			fmt.Println("Bad parameter:", os.Args[2])
		} else {
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
	//	fmt.Println("update:")
	if len(os.Args) != 5 {
		prtUsage()
	} else {
		var id int64
		if _, err := fmt.Sscanf(os.Args[2], "%d", &id); err != nil {
			fmt.Println("Bad parameter:", os.Args[2])
		} else {
			doUpdate(id, os.Args[3], os.Args[4])
		}
	}
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
	nArg := len(os.Args)
	if nArg < 3 || nArg > 4 {
		prtUsage()
	} else {
		var id int64
		if _, err := fmt.Sscanf(os.Args[2], "%d", &id); err != nil {
			fmt.Println("Bad parameter:", os.Args[2])
			return
		}
		if nArg == 4 {
			doGetFile(id, os.Args[3])
		} else {
			doCloneProj(id)
		}
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

		case "Search":
			fallthrough
		case "-S":
			searchCSProj()

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
