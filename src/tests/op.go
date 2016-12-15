// process net protocol, object operation

package main

import (
"fmt"
"errors"
"strings"
"encoding/json"
"os/exec"
"net"
"os"
"io"
"bufio"
)

var rsvr string ="123.206.55.31"
var rport string =":8192"

type PROJINFO struct{
	Id int64
	Title string
	Atime string // always use database updatetime
	Descr string
	Conclude string
	Path string
	IsDir bool
	Size int64
}

func (info* PROJINFO)dumpInfo(){
	fmt.Println("Id=",info.Id)
	fmt.Println("Title=",info.Title)
	fmt.Println("Atime=",info.Atime)
	fmt.Println("Descr=",info.Descr)
	fmt.Println("Conclude=",info.Conclude)
	fmt.Println("Path=",info.Path)
//	fmt.Println("IsDir=",info.IsDir)
	fmt.Println("Size=",info.Size)
	fmt.Println("")
}

func (info* PROJINFO)remoteCreate()(int64 , error){
/*
Client side:
 	-> Create \n
	-> json data \n
<- OK\n
 	-> binary file data
<- SUCCESS\n
<- json data \n
*/
	conn,err:=net.Dial("tcp",rsvr+rport)
	if err!=nil{
		return -1,err
	}
	defer conn.Close()

	rb:=bufio.NewReader(conn)
	// tar file, and get size 
	tmpfile:="/tmp/"+"proj.tgz"
	exec.Command("tar","czvf",tmpfile,info.Path).Run()
	st,_:=os.Stat(tmpfile)
	info.Size=st.Size()
	defer  os.Remove(tmpfile)
	if err:=info.confirm();err!=nil{
		return -1,err
	}
// parse path : first replace // with /, then remove anything before ../
	path:=info.Path
	for{
		sp:=strings.Replace(path,"//","/",-1)
		if sp!=path{
			path=sp
		}else{
			info.Path=sp
			break
		}
	}
	strs:=strings.Split(info.Path,"../")
	info.Path=strs[len(strs)-1]

	// send object by json, then send file
	var ctext string
	if obj,err:=json.Marshal(info);err!=nil{
		return -1,errors.New("struct json marshal error")
	}else{
		ctext=("Create\n")+string(obj)+"\n"
	}

	if _,err:=conn.Write([]byte(ctext));err!=nil{
		return -1,errors.New("send \"Create\" message error")
	}
//	buf:=make([]byte,1024,1024)
	buf,_,err:=rb.ReadLine()
	if err!=nil{
		return -1,errors.New("receive data error")
	}
	if string(buf) !="OK"{
		return -1,errors.New("remote server refused:"+string(buf))
	}

// copy file
	rd,_:=os.Open(tmpfile)
	io.CopyN(conn,rd,info.Size)
	rd.Close()
//	os.Remove(tmpfile)

	result,_,err:=rb.ReadLine()
	if(err!=nil){
		return -1,errors.New("Create--receive created data error")
	}
	if string(result)!="SUCCESS" {
		return -1,errors.New("receive failed:"+string(result))
	}
	buf,_,_=rb.ReadLine()
	if err:=json.Unmarshal(buf,info); err!=nil{
		return -1,errors.New("Create--resolve remote data error")
	}
	info.dumpInfo()
	return info.Id,nil
}

func(info* PROJINFO)confirm() error{
	fmt.Printf("Title: %s\nDescription: %s\nConclusion: %s\nPath:%s\nSize:%d\nAre you sure?(yes)",info.Title,info.Descr,info.Conclude,info.Path,info.Size)
	c:="yes"
	fmt.Scanf("%s",&c)
	if c=="yes"{
		return nil
	}else{
		return errors.New("user give up")
	}
}

func (info* PROJINFO)scanInfo(){
	rd:=bufio.NewReader(os.Stdin)
	fmt.Printf("Title (default: %s):",info.Title)
	bt,_,_:=rd.ReadLine()
	if len(bt)!=0{
		info.Title=string(bt)
	}
	fmt.Printf("Description (default: %s)",info.Descr)
	bt,_,_=rd.ReadLine()
	if len(bt)!=0{
		info.Descr=string(bt)
	}
//	fmt.Scanf("%s",&info.Descr)
	fmt.Printf("Conclusion (default: %s)",info.Conclude)
//	fmt.Scanf("%s",&info.Conclude)
	bt,_,_=rd.ReadLine()
	if len(bt)!=0{
		info.Conclude=string(bt)
	}
}

func doDel(id int64){
    conn,err:=net.Dial("tcp",rsvr+rport)
    if err!=nil{
		fmt.Println("connect to server error")
        return
    }
    defer conn.Close()
	conn.Write([]byte(fmt.Sprintf("Del\n%d\n",id)))
	rd:=bufio.NewReader(conn)
	for{
		line,_,err:=rd.ReadLine()
		if err!=nil{
			break
		}else{
			fmt.Println(string(line))
		}
	}
}

func doEdit(id int64){

/*
client side:
	-> Edit\n
	-> proj_id \n
<- OK\n
<- json data\n
	-> CANCEL\n  |  json_data\n
<- RESULT
*/
	conn,err:=net.Dial("tcp",rsvr+rport)
	if err!=nil{
		return
	}
	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("Edit\n%d\n",id)))
	rd:=bufio.NewReader(conn)
	line,_,err:=rd.ReadLine()
	if err!=nil{
		fmt.Println("Get Edit result error:",err)
		return
	}
	if string(line)!="OK"{
		fmt.Println("Find project error:",string(line))
		return
	}
	line,_,err=rd.ReadLine()
	if err!=nil{
		fmt.Println("Get remote project info error:",err)
		return
	}
	info:=new(PROJINFO)
	if err:=json.Unmarshal(line,info); err!=nil{
		fmt.Println("Edit--resolve remote data error")
		return
	}
	info.scanInfo()
	fmt.Printf("Project ID: %d\nTitle: %s\nDescription: %s\nConclusion: %s\nAre you sure?(yes)",info.Id,info.Title,info.Descr,info.Conclude)
	c:="yes"
	fmt.Scanf("%s",&c)
	ctext:="CANCEL\n"
	if c=="yes"{
		if obj,err:=json.Marshal(info);err!=nil{
			fmt.Println ("Edit client--Parse object error",err)
		}else{
			ctext=string(obj)+"\n"
			conn.Write([]byte(ctext))
			line,_,err=rd.ReadLine()
			if err!=nil{
				fmt.Println("Get edit result error:",err)
			}else{
				fmt.Println(string(line))
			}
			return
		}
	}
	fmt.Println("Edit canceled.")
	conn.Write([]byte(ctext))
}

func doList(){
    conn,err:=net.Dial("tcp",rsvr+rport)
    if err!=nil{
		fmt.Println("connect to server error")
        return
    }
    defer conn.Close()
	conn.Write([]byte("List\n"))
	rb:=bufio.NewReader(conn)
	line,_,_:=rb.ReadLine()
	var nObj int64
	if _,err:=fmt.Sscanf(string(line),"%d",&nObj);err!=nil{
		fmt.Println("Parse obj number error")
		return
	}
	obj:=new(PROJINFO)
	var i int64
	for i=0;i<nObj;i++{
		line,_,err=rb.ReadLine()
		if err!=nil{
			fmt.Println("Get remote data error:",err.Error())
			break;
		}
		if err:=json.Unmarshal(line,obj); err!=nil{
			fmt.Println("Resolve obj error:\n",string(line),"\n",err)
		}else{
			obj.dumpInfo()
		}
	}
}

func createInfo(path string, isdir bool) *PROJINFO{
	info:=&PROJINFO{-1,"noname","atime","No comment.","No explain,obviously.",path,isdir,0}
	info.scanInfo()
	return info
}

func ParseInput(conn net.Conn){
	input:=bufio.NewReader(os.Stdin)
	for{
		line,_:=input.ReadSlice('\n')
		if _,err:=conn.Write(line);err!=nil{
			break
		}
	}
}

func doRun(id int64, args []string){
    conn,err:=net.Dial("tcp",rsvr+rport)
    if err!=nil{
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
	if args==nil{
		conn.Write([]byte(fmt.Sprintf("%d 0\n",id)))
	}else{
		nArg:=len(args)
		conn.Write([]byte(fmt.Sprintf("%d %d\n",id,nArg)))
		for i:=0;i<nArg;i++{
			conn.Write([]byte(args[i]+"\n"))
		}
	}
	//conn.Write([]byte(fmt.Sprintf("Run\n%d",id)))
	rd:=bufio.NewReader(conn)
	// concurrent process intput and output
	go ParseInput(conn)
	for{
		line,longline,err:=rd.ReadLine()
		if err!=nil{
			break
		}else{
			if longline{
				fmt.Print(string(line))
			}else{
				fmt.Println(string(line))
			}
		}
	}
}

func doCreate(path string,isdir bool) (int64,error){
	// create object first
	info:=createInfo(path,isdir)
	return info.remoteCreate()
}

