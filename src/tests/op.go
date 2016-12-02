package main

import (
"fmt"
"errors"
"encoding/json"
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
	IsDir string
}



func (info* PROJINFO)remoteCreate()(int , erro){
	// send object by json, then send file
	if obj,err:=json.Marshal(info);err!=nil{
		return -1,errors.New("struct json marshal error")
	}
	if conn,err:=net.Dial("tcp",rsvr+rport); err!=nil{
		return -1,err
	}
	ctext:=[]byte("Create\n")+obj
	if _,err:=conn.Write(ctext);err!=nil{
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
	PROJINFO info:=&PROJINFO{-1,"noname","atime","No comment.","No explain,obviously.",Path,Isdir}
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

