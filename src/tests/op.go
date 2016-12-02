package main

import (
"fmt"
"errors"
"encoding/json"
"os/exec"
"net"
"os"
"io"
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
}



func (info* PROJINFO)remoteCreate()(int , error){
	// send object by json, then send file
	var ctext string
	if obj,err:=json.Marshal(info);err!=nil{
		return -1,errors.New("struct json marshal error")
	}else{
		ctext=("Create\n")+string(obj)
	}

	conn,err:=net.Dial("tcp",rsvr+rport)
	if err!=nil{
		return -1,err
	}
	if _,err:=conn.Write([]byte(ctext));err!=nil{
		conn.Close()
		return -1,errors.New("send \"Create\" message error")
	}
	buf:=make([]byte,1024,1024)
	if _,err:=conn.Read(buf);err!=nil{
		conn.Close()
		return -1,errors.New("receive data error")
	}
	if string(buf) !="OK"{
		conn.Close()
		return -1,errors.New("remove server refused:"+string(buf))
	}
	// tar file, and send
	tmpfile:="/tmp/"+info.Path+".tgz"
	cmd:=fmt.Sprintf("tar czvf %s %s",tmpfile,info.Path)
	exec.Command(cmd)
// copy file
	rd,_:=os.Open(tmpfile)
	io.Copy(conn,rd)
	os.Remove("/tmp/"+info.Path+".tgz")
// refill object by json data
	if _,err:=conn.Read(buf);err!=nil{
		conn.Close()
		return -1,errors.New("receive created data error")
	}
	conn.Close()
	if err:=json.Unmarshal(buf,info); err!=nil{
		return -1,errors.New("resolve remote data error")
	}
	return info.Id,nil
}

func(info* PROJINFO)confirm() error{
	fmt.Printf("Title: %s\nDescription: %s\nConclusion: %s\nPath:%s\nAre you sure?(yes)",info.Title,info.Descr,info.Conclude,info.Path)
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
	fmt.Scanf("%s",&info.Descr)
	fmt.Printf("Conclusion (default: %s)",info.Conclude)
	fmt.Scanf("%s",&info.Conclude)
}

func doList(){
	fmt.Println("Do list")
	// query remote mariadb directly (first)
}

func createInfo(path string, isdir bool) *PROJINFO{
	info:=&PROJINFO{-1,"noname","atime","No comment.","No explain,obviously.",path,isdir}
	info.scanInfo()
	return info
}

func doCreate(path string,isdir bool) (int,error){
	// create object first
	info:=createInfo(path,isdir)
	if err:=info.confirm();err==nil{
		return info.remoteCreate()
	}else{
		return -1,err
	}
}

