package main

import (
"fmt"
"net"
//"sync"
"os/exec"
"time"
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
	switch string(command){
	case "Create":
		proj:=new(backend.PROJINFO)
		buf:=make([]byte,4096,4096)
		if  len,err:=rd.Read(buf); err!=nil{
			fmt.Println("Read create parameters error:",err)
			return
		}else if err:=json.Unmarshal(buf[:len],proj);err!=nil{
			fmt.Println("Resolve create data error:",string(buf))
			conn.Write([]byte("ERROR "+err.Error()))
			return
		}
		if err:=proj.CreateInDB(); err!=nil{
			fmt.Println("Create in database error:",err)
			conn.Write([]byte("ERROR "+err.Error()))
			return
		}
		conn.Write([]byte("OK"))
		exec.Command("mkdir", "-p", fmt.Sprintf("/opt/testssvr/%d",proj.Id)).Run()
		if crfile,err:=os.Create(fmt.Sprintf("/opt/testssvr/%d/proj.tgz",proj.Id));err==nil{
			io.CopyN(crfile,conn,proj.Size)
			crfile.Close()
			obj,_:=json.Marshal(proj)
			ret:="SUCCESS\n"+string(obj)
			conn.Write([]byte(ret))
		}else{
			fmt.Println("Create file error:",err)
			conn.Write([]byte("ERROR "+err.Error()))
			return
		}
	case "List":
		projs,err:=backend.ListProj()
		if err!=nil{
			conn.Write([]byte("ERROR "+err.Error()))
			return
		}
		objs:=len(projs)
		conn.Write([]byte(fmt.Sprintf("%d\n",objs)))
		for i:=0;i<objs;i++{
			objbuf,_:=json.Marshal(projs[i])
			line:=string(objbuf)+"\n"
			conn.Write([]byte(line))
		}
	case "Run":
		bufid:=make([]byte,10,10)
		if nRd,err:=rd.Read(bufid);err!=nil{
			fmt.Println("Read id error",err)
//			conn.Write([]byte("ERROR "+err.Error())) // need not,since read error already
			return
		}else{
			var nID int64
			_,err:=fmt.Sscanf(string(bufid[:nRd]),"%d",&nID)
			cmd,err:=backend.RunID(nID) // cmd.Start
			// concurrent input and output
			if err!=nil{
				fmt.Println("Run command error:",err)
				conn.Write([]byte("ERROR "+err.Error()))
				return
			}
			//outch:=make(chan string,5)
			//waitch:=make(chan int,1)
			outp,_:=cmd.StdoutPipe()
			inputp,_:=cmd.StdinPipe()
			go sendOutput(outp,conn)
			go getInput(inputp,rd)
			if err:=cmd.Start();err!=nil{
				fmt.Println("Start program error")
				return
			}
			cmd.Wait()
			time.Sleep(500*time.Millisecond)
		}
	default:
		fmt.Println("Unknown command:",command)
	}
}

func sendOutput(outp io.ReadCloser,conn net.Conn){
	defer outp.Close()
	buf:=make([]byte,1024,4096)
	for{
		n,err:=outp.Read(buf)
		if err!=nil{
			//fmt.Println("Read pipe over:",err)
			break
		}else{
			if _,err:=conn.Write(buf[:n]);err!=nil{
			//	fmt.Println("Send output to client failed:",err)
				break
			}
		}
	}
//	fmt.Println("End output routin")
}

func getInput(inp io.WriteCloser, rd *bufio.Reader){
	defer inp.Close()
	for{
		if line,err:=rd.ReadSlice('\n');err!=nil{
//			fmt.Println("Get remote input failed:",err)
			break
		}else{
			inp.Write(line)
		}
	}
//	fmt.Println("End input routin")
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
