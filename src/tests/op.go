// process net protocol, object operation

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var rsvr string = "123.206.55.31"
var rport string = ":8192"

type PROJINFO struct {
	Id       int64
	Title    string
	Atime    string // always use database updatetime
	Descr    string
	Conclude string
	Path     string
	IsDir    bool
	Size     int64
}

func (info *PROJINFO) dumpInfo() string {
	return fmt.Sprintf("ProjectID = [%d]\nTitle = [%s]\nModified = [%s]\nDescription = [%s]\nConclusion = [%s]\nPath = [%s]\n",
		info.Id, info.Title, info.Atime, info.Descr, info.Conclude, info.Path)
}

func (info *PROJINFO) remoteCreate() (int64, error) {
	/*
	   Client side:
	    	-> Create \n
	   	-> json data \n
	   <- OK\n
	    	-> binary file data
	   <- SUCCESS\n
	   <- json data \n
	*/
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	rb := bufio.NewReader(conn)
	// tar file, and get size
	tmpfile := "/tmp/" + "proj.tgz"
	exec.Command("tar", "czvf", tmpfile, info.Path).Run()
	st, _ := os.Stat(tmpfile)
	info.Size = st.Size()
	defer os.Remove(tmpfile)
	if err := info.confirm(); err != nil {
		return -1, err
	}
	// parse path : first replace // with /, then remove anything before ../
	path := info.Path
	for {
		sp := strings.Replace(path, "//", "/", -1)
		if sp != path {
			path = sp
		} else {
			info.Path = sp
			break
		}
	}
	strs := strings.Split(info.Path, "../")
	info.Path = strs[len(strs)-1]

	// send object by json, then send file
	var ctext string
	if obj, err := json.Marshal(info); err != nil {
		return -1, errors.New("struct json marshal error")
	} else {
		ctext = ("Create\n") + string(obj) + "\n"
	}

	if _, err := conn.Write([]byte(ctext)); err != nil {
		return -1, errors.New("send \"Create\" message error")
	}
	buf, _, err := rb.ReadLine()
	if err != nil {
		return -1, errors.New("receive data error")
	}
	if string(buf) != "OK" {
		return -1, errors.New("remote server refused:" + string(buf))
	}

	// copy file
	rd, _ := os.Open(tmpfile)
	defer rd.Close()
	if _, err := io.CopyN(conn, rd, info.Size); err != nil {
		fmt.Println("Copy file to server error:", err)
		return -1, errors.New("Create--Copy file error")
	}

	result, _, err := rb.ReadLine()
	if err != nil {
		return -1, errors.New("Create--receive created data error")
	}
	if string(result) != "SUCCESS" {
		return -1, errors.New("receive failed:" + string(result))
	}
	buf, _, _ = rb.ReadLine()
	if err := json.Unmarshal(buf, info); err != nil {
		return -1, errors.New("Create--resolve remote data error")
	}
	return info.Id, nil
}

func (info *PROJINFO) confirm() error {
	fmt.Printf("Title: %s\nDescription: %s\nConclusion: %s\nPath:%s\nSize:%d\nAre you sure?(yes)", info.Title, info.Descr, info.Conclude, info.Path, info.Size)
	c := "yes"
	fmt.Scanf("%s", &c)
	if c == "yes" {
		return nil
	} else {
		return errors.New("user give up")
	}
}

func (info *PROJINFO) scanInfo() {
	rd := bufio.NewReader(os.Stdin)
	fmt.Printf("Title (default: %s):", info.Title)
	bt, _, _ := rd.ReadLine()
	if len(bt) != 0 {
		info.Title = string(bt)
	}
	fmt.Printf("Description (default: %s)", info.Descr)
	bt, _, _ = rd.ReadLine()
	if len(bt) != 0 {
		info.Descr = string(bt)
	}
	//	fmt.Scanf("%s",&info.Descr)
	fmt.Printf("Conclusion (default: %s)", info.Conclude)
	//	fmt.Scanf("%s",&info.Conclude)
	bt, _, _ = rd.ReadLine()
	if len(bt) != 0 {
		info.Conclude = string(bt)
	}
}

func doDel(id int64) {
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return
	}
	defer conn.Close()
	conn.Write([]byte(fmt.Sprintf("Del\n%d\n", id)))
	rd := bufio.NewReader(conn)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			break
		} else {
			fmt.Println(string(line))
		}
	}
}

func doEdit(id int64) {
	/*
	   client side:
	   	-> Edit\n
	   	-> proj_id \n
	   <- OK\n
	   <- json data\n
	   	-> CANCEL\n  |  json_data\n
	   <- RESULT
	*/
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("Edit\n%d\n", id)))
	rd := bufio.NewReader(conn)
	line, _, err := rd.ReadLine()
	if err != nil {
		fmt.Println("Get Edit result error:", err)
		return
	}
	if string(line) != "OK" {
		fmt.Println("Find project error:", string(line))
		return
	}
	line, _, err = rd.ReadLine()
	if err != nil {
		fmt.Println("Get remote project info error:", err)
		return
	}
	info := new(PROJINFO)
	if err := json.Unmarshal(line, info); err != nil {
		fmt.Println("Edit--resolve remote data error")
		return
	}
	info.scanInfo()
	fmt.Printf("Project ID: %d\nTitle: %s\nDescription: %s\nConclusion: %s\nAre you sure?(yes)", info.Id, info.Title, info.Descr, info.Conclude)
	c := "yes"
	fmt.Scanf("%s", &c)
	ctext := "CANCEL\n"
	if c == "yes" {
		if obj, err := json.Marshal(info); err != nil {
			fmt.Println("Edit client--Parse object error", err)
		} else {
			ctext = string(obj) + "\n"
			conn.Write([]byte(ctext))
			line, _, err = rd.ReadLine()
			if err != nil {
				fmt.Println("Get edit result error:", err)
			} else {
				fmt.Println(string(line))
			}
			return
		}
	}
	fmt.Println("Edit canceled.")
	conn.Write([]byte(ctext))
}

func getBrowseContent(id int64) []string {
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return nil
	}
	defer conn.Close()
	projsrc := make([]string, 0, 50)
	conn.Write([]byte(fmt.Sprintf("Browse\n%d\n", id)))
	rd := bufio.NewReader(conn)
	for {
		line, _, err := rd.ReadLine()
		if err != nil {
			break
		} else {
			//	fmt.Println(string(line))
			projsrc = append(projsrc, string(line))
		}
	}
	return projsrc
}

func doBrowse(id int64) {
	projsrc := getBrowseContent(id)
	if projsrc != nil {
		for _, line := range projsrc {
			fmt.Println(line)
		}
	}
}

func doUpdate(id int64, projfile string, localfile string) {
	/*
	   client side:
	   	-> Update\n
	   	-> ID \n
	   	-> projfile \n
	   <- nOrgFileSize\n | ERROR No such file\n
	   	-> nFileSize \n | CANCEL \n
	   	->RawFile
	   <- OK | ERROR
	*/

	fLocal, err := os.Stat(localfile)
	if err != nil {
		fmt.Println("Can't find file: ", localfile)
		return
	} else if fLocal.IsDir() {
		fmt.Println("Support update a file only.")
		return
	}
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return
	}
	defer conn.Close()
	conn.Write([]byte(fmt.Sprintf("Update\n%d\n%s\n", id, projfile)))
	rd := bufio.NewReader(conn)
	line, _, err := rd.ReadLine()
	reply := string(line)
	if strings.Contains(reply, "ERROR") {
		fmt.Println("Update error:\n", reply)
		return
	}
	var rsize int64 = 0
	lsize := fLocal.Size()
	if _, err := fmt.Sscanf(reply, "%d", &rsize); err != nil {
		fmt.Println("Resolve remote file size error.")
		return
	}
	fmt.Printf("Replace remote file: %s(%d bytes) by local file %s(%dbytes)? (yes)", projfile, rsize, localfile, lsize)
	cf := "yes"
	fmt.Scanf("%s", &cf)
	if cf != "yes" {
		fmt.Println("Update cancelled!")
		conn.Write([]byte("CANCEL\n"))
	} else {
		conn.Write([]byte(fmt.Sprintf("%d\n", lsize)))
		sfile, _ := os.Open(localfile)
		defer sfile.Close()
		if _, err := io.CopyN(conn, sfile, lsize); err != nil {
			fmt.Println("Upload file failed")
			return
		}
		if line, _, err := rd.ReadLine(); err != nil {
			fmt.Println("Get update result error: ", err)
		} else if string(line) != "OK" {
			fmt.Println("Update file error: ", string(line))
		} else {
			fmt.Println("Update file ok!")
		}
	}
}

func doGetFile(id int64, path string) {
	doGetFileContent(id, path, os.Stdout)
	fmt.Println("")
}

func doGetFileContent(id int64, path string, dst io.Writer) error {
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return errors.New("connect to server")
	}
	defer conn.Close()
	conn.Write([]byte(fmt.Sprintf("Get\n%d\n%s\n", id, path)))
	_, err = io.Copy(dst, conn)
	return err
}

func doCloneProj(id int64) {
	projsrc := getBrowseContent(id)
	if projsrc == nil {
		return
	}
	if len(projsrc) == 1 && strings.HasPrefix(projsrc[0], "ERROR ") {
		fmt.Println(projsrc[0])
		return
	}
	projpath := fmt.Sprintf("testsproj-%d", id)
	if _, err := os.Stat(projpath); err == nil {
		fmt.Println(projpath + " already exists")
		return
	}
	if err := os.Mkdir(projpath, 0755); err != nil {
		fmt.Println(err)
		return
	}
	os.Chdir(projpath)
	patdir, _ := regexp.Compile("^\\[dir\\]\\s+(\\S+.*)$")
	patfile, _ := regexp.Compile("^\\s+\\[(\\d+)\\|(\\d+)\\]\\s+(\\S+.*)$")
	nAll := len(projsrc)
	nCur := 0
	for _, line := range projsrc {
		nCur++
		prog := fmt.Sprintf("(%d/%d)\t", nCur, nAll)
		if dirname := patdir.FindStringSubmatch(line); dirname != nil {
			fmt.Print(prog + dirname[1] + "...")
			os.Mkdir(dirname[1], 0755) // since parent create successfully, do not check error here
			fmt.Println("OK")
		} else if file := patfile.FindStringSubmatch(line); file != nil {
			//fmt.Println(filename[1],filename[2],filename[3])
			mode, _ := strconv.ParseInt(file[2], 8, 32)
			fmode := os.FileMode(mode)
			if cfile, err := os.OpenFile(file[3], os.O_CREATE|os.O_WRONLY, fmode); err != nil {
				fmt.Println(err)
				return
			} else {
				// os.Remove maybe used, so defer cfile.Close is avoid here
				fmt.Print(prog + file[3] + "...")
				err := doGetFileContent(id, file[3], cfile)
				if err != nil {
					fmt.Println(err)
					cfile.Close()
					os.Remove(file[3])
					fmt.Println("failed")
					return
				} else {
					cfile.Close()
					fmt.Println("OK")
				}
			}
		}
	}
	fmt.Println("Project sources download successfully in ./" + projpath)
}

func doList() {
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return
	}
	defer conn.Close()
	conn.Write([]byte("List\n"))
	rb := bufio.NewReader(conn)
	line, _, _ := rb.ReadLine()
	var nObj int64
	if _, err := fmt.Sscanf(string(line), "%d", &nObj); err != nil {
		fmt.Println("Parse obj number error")
		return
	}
	obj := new(PROJINFO)
	var i int64
	for i = 0; i < nObj; i++ {
		line, _, err = rb.ReadLine()
		if err != nil {
			fmt.Println("Get remote data error:", err.Error())
			return
		}
		if err := json.Unmarshal(line, obj); err != nil {
			fmt.Println("Resolve obj error:\n", string(line), "\n", err)
		} else {
			fmt.Println(obj.dumpInfo())
		}
	}
	fmt.Println(nObj, "projects listed.\n")
	ListPlugins()
}

func ListPlugins() {
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return
	}
	defer conn.Close()
	conn.Write([]byte("Plugin\n"))
	rd := bufio.NewReader(conn)
	if line, _, err := rd.ReadLine(); err != nil {
		return
	} else {
		if nPlugin, _ := strconv.Atoi(string(line)); nPlugin != 0 {
			fmt.Println("The source code of following language(s) may be run directly:")
			for i := 0; i < nPlugin; i++ {
				if line, _, err := rd.ReadLine(); err == nil {
					fmt.Print(string(line) + "  ")
				} else {
					break
				}
			}
			fmt.Println("")
		}
	}
}

func createInfo(path string, isdir bool) *PROJINFO {
	info := &PROJINFO{-1, "noname", "atime", "No comment.", "No explain, obviously.", path, isdir, 0}
	info.scanInfo()
	return info
}

func ParseInput(conn net.Conn) {
	input := bufio.NewReader(os.Stdin)
	for {
		line, _, _ := input.ReadLine()
		line = append(line, '\n')
		if _, err := conn.Write(line); err != nil {
			break
		}
	}
}

func doSearch(keywords []string) {
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return
	}
	defer conn.Close()
	/*
	   search client:
	   	-> Search\n
	   	-> nArgs
	   	-> keyword1\n
	   	-> keyword2\n
	   	-> ...
	   	-> keywordn\n
	   <- nProjs
	   <- Result
	*/
	conn.Write([]byte(fmt.Sprintf("Search\n%d\n", len(keywords))))
	for _, arg := range keywords {
		conn.Write([]byte(arg + "\n"))
	}
	rd := bufio.NewReader(conn)
	var nObj int64 = 0
	if objline, _, err := rd.ReadLine(); err != nil {
		fmt.Println("Get project's number error:", err)
		return
	} else {
		fmt.Sscanf(string(objline), "%d", &nObj)
	}
	obj := new(PROJINFO)
	var i int64
	for i = 0; i < nObj; i++ {
		line, _, err := rd.ReadLine()
		if err != nil {
			fmt.Println("Get remote data error:", err.Error())
			break
		}
		if err := json.Unmarshal(line, obj); err != nil {
			fmt.Println("Resolve obj error:\n", string(line), "\n", err)
		} else {
			objinfo := obj.dumpInfo()
			for _, kw := range keywords {
				if strings.ToLower(runtime.GOOS) == "windows" {
					objinfo = strings.Replace(objinfo, kw, ">"+kw+"<", -1)
				} else {
					objinfo = strings.Replace(objinfo, kw, "\033[7m"+kw+"\033[0m", -1)
				}
			}
			fmt.Println(objinfo)
		}
	}
}

func doRun(id int64, args []string) {
	conn, err := net.Dial("tcp", rsvr+rport)
	if err != nil {
		fmt.Println("connect to server error")
		return
	}
	defer conn.Close()
	/*
		Run\n
		id nArgs\n
		arg1\n
		arg2\n
		...
		argn\n
	*/
	conn.Write([]byte("Run\n"))
	if args == nil {
		conn.Write([]byte(fmt.Sprintf("%d 0\n", id)))
	} else {
		nArg := len(args)
		conn.Write([]byte(fmt.Sprintf("%d %d\n", id, nArg)))
		for i := 0; i < nArg; i++ {
			conn.Write([]byte(args[i] + "\n"))
		}
	}
	//conn.Write([]byte(fmt.Sprintf("Run\n%d",id)))
	rd := bufio.NewReader(conn)
	// concurrent process intput and output
	go ParseInput(conn)
	for {
		line, longline, err := rd.ReadLine()
		if err != nil {
			break
		} else {
			if longline {
				fmt.Print(string(line))
			} else {
				fmt.Println(string(line))
			}
		}
	}
}

func doCreate(path string, isdir bool) (int64, error) {
	// create object first
	info := createInfo(path, isdir)
	return info.remoteCreate()
}
