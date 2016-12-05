package main

import (
"fmt"
"net"
"os/exec"
"bufio"
"io"
"os"
"backend"
"encoding/json"
)

func procConn(conn net.Conn){
	defer conn.Close()
	rd:=bufio.NewReader(conn)
	command,_,err:=rd.ReadLine()
	if err!=nil{
		fmt.Println("Read command error:",err)
		return
	}
	proj:=new(backend.PROJINFO)
	switch string(command){
	case "Create":
		buf:=make([]byte,4096,4096)
		if  _,err:=rd.Read(buf); err!=nil{
			fmt.Println("Read create parameters error:",err)
			return
		}
		if err:=json.Unmarshal(buf,proj);err!=nil{
			fmt.Println("Resove create data error:",string(buf))
			conn.Write([]byte("ERROR "+err.Error()))
			return
		}
		if err:=proj.CreateInDB(); err!=nil{
			fmt.Println("Create in database error:",err)
			conn.Write([]byte("ERROR "+err.Error()))
			return
		}
		conn.Write([]byte("OK"))
		exec.Command(fmt.Sprintf("mkdir -p /opt/testssvr/%d",proj.Id))
		if crfile,err:=os.Create(fmt.Sprintf("/opt/testssvr/%d/proj.tgz",proj.Id));err==nil{
			io.Copy(crfile,conn)
			crfile.Close()
			obj,_:=json.Marshal(proj)
			ret:="SUCCESS\n"+string(obj)
			conn.Write([]byte(ret))
		}else{
			fmt.Println("Create file error:",err)
			conn.Write([]byte("ERROR "+err.Error()))
			return
		}
	default:
		fmt.Println("Unknown command:",command)
	}
}

func main(){
	lisn,err:=net.Listen("tcp",":8192")
	if err!=nil{
		fmt.Println("Server listen error:",err)
		return
	}
	defer lisn.Close()
	for{
		conn,err:=lisn.Accept()
		if err!=nil{
			fmt.Println("Server accept error:",err)
			return
		}
		go procConn(conn)
	}
}
