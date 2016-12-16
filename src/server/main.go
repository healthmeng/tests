//  process net protocol related io operations
package main

import (
	"fmt"
	"net"
	//"sync"
	//"os/exec"
	"time"
	"backend"
	"bufio"
	"encoding/json"
	"io"
	"os"
)

type RemoteIO struct {
	rdr *bufio.Reader
	wtr net.Conn
}

func procConn(conn net.Conn) {
	defer conn.Close()
	rd := bufio.NewReader(conn)
	command, _, err := rd.ReadLine()
	if err != nil {
		fmt.Println("Read command error:", err)
		return
	}
	switch string(command) {
	/*
	   Create procedure:
	   Server side:
	   <- Create \n
	   <- JsonData \n
	   	-> OK\n
	   <- binary file data
	   	-> SUCCESS\n
	   	-> JsonData \n
	*/
	case "Create":
		proj := new(backend.PROJINFO)
		//		buf:=make([]byte,4096,4096)
		if buf, _, err := rd.ReadLine(); err != nil {
			fmt.Println("Read create parameters error:", err)
			return
		} else if err := json.Unmarshal(buf, proj); err != nil {
			fmt.Println("Resolve create data error:", string(buf))
			conn.Write([]byte("ERROR: " + err.Error()))
			return
		}
		conn.Write([]byte("OK\n"))
		tmpfile:=fmt.Sprintf("/tmp/proj-%d",time.Now().UnixNano())
		if crfile, err := os.Create(tmpfile); err == nil {
			defer os.Remove(tmpfile)
			size,recverr:=io.CopyN(crfile, conn, proj.Size)
			crfile.Close()
			if recverr!=nil || size!=proj.Size{
				fmt.Println("Receive file error")
				conn.Write([]byte("Receive file error"))
				return
			}
			if err := proj.CreateInDB(); err != nil {
				fmt.Println("Create in database error:", err)
				conn.Write([]byte("ERROR: " + err.Error()))
				return
			}
		/*	projdir:=getProjDir(proj.Id)
			exec.Command("mkdir", "-p", projdir).Run()
			if err:=exec.Command("tar","xzvf",tmpfile,"-C",projdir).Run();err!=nil{*/
			if err=proj.InitDir(tmpfile);err!=nil{
				fmt.Println("Bad tgz file, can't uncompress: ",err)
				backend.DelProj(proj.Id)
				conn.Write([]byte("ERROR: project tarball is not a valid .tgz file"))
			}else{
				obj, _ := json.Marshal(proj)
				ret := "SUCCESS\n" + string(obj) + "\n"
				conn.Write([]byte(ret))
			}
		} else {
			fmt.Println("Create file error:", err)
			conn.Write([]byte("ERROR: " + err.Error()))
		}

	case "List":
		projs, err := backend.ListProj()
		if err != nil {
			conn.Write([]byte("ERROR " + err.Error()))
			return
		}
		objs := len(projs)
		conn.Write([]byte(fmt.Sprintf("%d\n", objs)))
		for i := 0; i < objs; i++ {
			objbuf, _ := json.Marshal(projs[i])
			line := string(objbuf) + "\n"
			conn.Write([]byte(line))
		}

	case "Del":
		if bufid, _, err := rd.ReadLine(); err != nil {
			fmt.Println("Del proj:read id error--", err)
		} else {
			var nID int64
			fmt.Sscanf(string(bufid), "%d", &nID)
			if err := backend.DelProj(nID); err != nil {
				conn.Write([]byte("Del failed:" + err.Error()))
			} else {
				conn.Write([]byte("Del success!"))
			}
		}

	case "Browse":
		if bufid,_,err:=rd.ReadLine();err!=nil{
			fmt.Println("Browse proj: read id error--",err)
		} else{
			var nID int64
			fmt.Sscanf(string(bufid),"%d",&nID)
			if files,err:=backend.BrowseProj(nID);err!=nil{
				fmt.Println("Browse proj failed:",err)
				conn.Write([]byte("Browse proj error:"+err.Error()))
			}else{
				for _,line:=range(files){
					conn.Write([]byte(line+"\n"))
				}
			}
		}
		/*
		   Edit procedure:
		   server side:
		   <- Edit\n
		   <- proj_id \n
		   	-> OK\n
		   	-> proj json data\n
		   <- CANCEL\n (close) |  json_data\n
		   	-> RESULT(if get jsondata)
		*/

	case "Edit":
		if bufid, _, err := rd.ReadLine(); err != nil {
			fmt.Println("Edit proj:read id error", err)
		} else {
			var nID int64
			fmt.Sscanf(string(bufid), "%d", &nID)
			if proj, err := backend.LookforID(nID); err != nil {
				conn.Write([]byte("Can't find id in db:" + err.Error()))
			} else {
				conn.Write([]byte("OK\n"))
				obj, _ := json.Marshal(proj)
				ret := string(obj) + "\n"
				conn.Write([]byte(ret))
				buf, _, err := rd.ReadLine()
				if err != nil {
					fmt.Println("Get edit response error:", err)
					return
				}
				if string(buf) != "CANCEL" {
					if err := json.Unmarshal(buf, proj); err != nil {
						fmt.Println("Resolve create data error:", string(buf))
						conn.Write([]byte("ERROR " + err.Error()))
					} else {
						if err = proj.UpdateDB(); err != nil {
							conn.Write([]byte("Update database failed:" + err.Error()))
						} else {
							conn.Write([]byte("Edit success!"))
						}
					}
				} // else remote finish connection
			}
		}

	case "Run":
		//bufid:=make([]byte,10,10)
		if bufid, _, err := rd.ReadLine(); err != nil {
			fmt.Println("Run proj:read id error", err)
			//			conn.Write([]byte("ERROR "+err.Error())) // need not,since read error already
			return
		} else {
			var nID int64
			var nParam int
			fmt.Sscanf(string(bufid), "%d%d", &nID, &nParam)
			params := make([]string, 0, 20)
			for i := 0; i < nParam; i++ {
				param, _, err := rd.ReadLine()
				if err != nil {
					fmt.Println("Get parameters failed.")
					conn.Write([]byte("Parameters error"))
					return
				}
				params = append(params, string(param))
			}
			rio := &RemoteIO{rdr: rd, wtr: conn}
			chout, err := backend.RunID(nID, rio, params) // cmd.Start,concurrent input and output
			if err != nil {
				fmt.Println("Run command error:", err)
				conn.Write([]byte("ERROR " + err.Error()))
			} else {
				<-chout
			}
		}

	default:
		fmt.Println("Unknown command:", command)
	}
}

func (r *RemoteIO) SendOutput(outp io.ReadCloser) {
	defer outp.Close()
	buf := make([]byte, 1024, 4096)
	for {
		n, err := outp.Read(buf)
		if err != nil {
			//fmt.Println("Read pipe over:",err)
			break
		} else {
			if _, err := r.wtr.Write(buf[:n]); err != nil {
				//	fmt.Println("Send output to client failed:",err)
				break
			}
		}
	}
}

func (r *RemoteIO) GetInput(inp io.WriteCloser) {
	defer inp.Close()
	for {
		if line, err := r.rdr.ReadSlice('\n'); err != nil {
			//			fmt.Println("Get remote input failed:",err)
			break
		} else {
			inp.Write(line)
		}
	}
	//	fmt.Println("End input routin")
}

func main() {
	lisn, err := net.Listen("tcp", ":8192")
	if err != nil {
		fmt.Println("Server listen error:", err)
		return
	}
	defer lisn.Close()
	for {
		conn, err := lisn.Accept()
		if err != nil {
			fmt.Println("Server accept error:", err)
			return
		}
		go procConn(conn)
	}
}
