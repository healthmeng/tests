package main

import (
"fmt"
"errors"
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
	// send object by json, then send file
	var ctext string
	if obj,err:=json.Marshal(info);err!=nil{
		return -1,errors.New("struct json marshal error")
	}else{
		ctext=("Create\n")+string(obj)
	}

	if _,err:=conn.Write([]byte(ctext));err!=nil{
		return -1,errors.New("send \"Create\" message error")
	}
	buf:=make([]byte,1024,1024)
	if _,err:=rb.Read(buf);err!=nil{
		return -1,errors.New("receive data error")
	}
	if string(buf[:2]) !="OK"{
		return -1,errors.New("remote server refused:"+string(buf))
	}

// copy file
	rd,_:=os.Open(tmpfile)
	io.CopyN(conn,rd,info.Size)
	rd.Close()
//	os.Remove(tmpfile)

	result,_,err:=rb.ReadLine()
	if(err!=nil){
		return -1,errors.New("receive created data error")
	}
	if string(result)!="SUCCESS" {
		return -1,errors.New("receive failed:"+string(result))
	}
	len,_:=rb.Read(buf)
	if err:=json.Unmarshal(buf[:len],info); err!=nil{
		return -1,errors.New("resolve remote data error")
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

func doList(){
    conn,err:=net.Dial("tcp",rsvr+rport)
    if err!=nil{
		fmt.Println("connect to server error")
        return
    }
    defer conn.Close()
fmt.Println("Start write")
	conn.Write([]byte("List\n"))
	rb:=bufio.NewReader(conn)
	line,_,_:=rb.ReadLine()
	var nObj int64
	if _,err:=fmt.Sscanf(string(line),"%d",&nObj);err!=nil{
		fmt.Println("Parse obj number error")
		return
	}
fmt.Println("Start list proj",nObj)
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
fmt.Println("Start dump")
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
		line,_,_:=input.ReadLine()
		if _,err:=conn.Write(line);err!=nil{
			break
		}
	}
}

func doRun(id int64){
    conn,err:=net.Dial("tcp",rsvr+rport)
    if err!=nil{
        fmt.Println("connect to server error")
        return
    }
    defer conn.Close()
	conn.Write([]byte(fmt.Sprintf("Run\n%d",id)))
	rd:=bufio.NewReader(conn)
	// concurrent process intput and output
	go ParseInput(conn)
	for{
		line,isline,err:=rd.ReadLine()
		if err!=nil{
			break
		}else{
			if isline{
				fmt.Println(string(line))
			}else{
				fmt.Print(string(line))
			}
		}
	}
}

func doCreate(path string,isdir bool) (int64,error){
	// create object first
	info:=createInfo(path,isdir)
	return info.remoteCreate()
}

