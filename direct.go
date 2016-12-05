package main

import (
"fmt"
"net"
"io"
"os"
"strings"
)
/*
func getret(cmd string)(ret string,err error){
	arg:=strings.Split(cmd," ")
	args:=arg[1:]
	c:=exec.Command(arg[0],args...)
	buf,err:=c.Output()
	if err!=nil{
		return "",err
	}else{
		return string(buf),nil
	}
}
*/
func main(){
/*	var cmd string
	reader:=bufio.NewReader(os.Stdin)
	line,_,_:=reader.ReadLine()
	cmd=string(line)
	arg:=strings.Split(cmd," ")
	args:=arg[1:]
	c:=exec.Command(arg[0],args...)
	buf,err:=c.Output()
	if err!=nil{
		fmt.Println("error:",err.Error())
	}else{
		fmt.Println(string(buf))
	}
*/
	lis,err:=net.Listen("tcp",":12345")
	if err!=nil{
		fmt.Println("listen error:",err)
		os.Exit(1)
	}
	for{
		conn,err:=lis.Accept()
		if err != nil{
			fmt.Println("Accept error:",err.Error())
			lis.Close()
			break
		}
	//	ch:=make(chan string,1)
		go comm(conn)
	}
}

func comm(conn net.Conn){
	fmt.Println("Accepted")
	buf:=make([]byte,4096,8192)	
	n,err:=conn.Read(buf)
	if (err!=nil) {
		fmt.Println("Read error:",err.Error())
		conn.Close()
		return
	}
/*	for{
	if buf[n-1]=='\n' {
		n--
		buf=buf[0:n]
	}else{
		break
	}
	}*/
	res:=strings.Trim(string(buf[:n]),"\n\r")
	fmt.Println(res+":")
/*	str,err:=getret(string(buf))
	if err==nil{
		conn.Write([]byte(str))
	}*/
	if file,err:=os.Open(res);err!=nil{
		conn.Write([]byte("file not found\n"))
	}else{
		io.Copy(conn,file)
		conn.Write([]byte("\n"))
	}
	conn.Close()
}
