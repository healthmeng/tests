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
	Id int
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
	fmt.Println("IsDir=",info.IsDir)
	fmt.Println("Size=",info.Size)
}

func (info* PROJINFO)remoteCreate()(int , error){

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
	io.Copy(conn,rd)
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
		fmt.Println(string(buf[:len]))
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
	fmt.Printf("Title (default: %s):",info.Title)
	fmt.Scanf("%s",&info.Title)
	fmt.Printf("Description (default: %s)",info.Descr)
	rd:=bufio.NewReader(os.Stdin)
	bt,_,_:=rd.ReadLine()
	info.Descr=string(bt)
//	fmt.Scanf("%s",&info.Descr)
	fmt.Printf("Conclusion (default: %s)",info.Conclude)
//	fmt.Scanf("%s",&info.Conclude)
	bt,_,_=rd.ReadLine()
	info.Conclude=string(bt)
}

func doList(){
	fmt.Println("Do list")
	// query remote mariadb directly (first)
}

func createInfo(path string, isdir bool) *PROJINFO{
	info:=&PROJINFO{-1,"noname","atime","No comment.","No explain,obviously.",path,isdir,0}
	info.scanInfo()
	return info
}

func doCreate(path string,isdir bool) (int,error){
	// create object first
	info:=createInfo(path,isdir)
	return info.remoteCreate()
}

