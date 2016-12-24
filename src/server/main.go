//  process net protocol related io operations
package main

import (
	"fmt"
	"net"
	//"sync"
	//"os/exec"
	"time"
	"strings"
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
	case "Create":
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
			size,recverr:=io.CopyN(crfile, rd, proj.Size)
			crfile.Close()
			if recverr!=nil || size!=proj.Size{
				fmt.Println("Receive file error")
				conn.Write([]byte("ERROR Receive file error"))
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
				conn.Write([]byte("ERROR Del failed:" + err.Error()))
			} else {
				conn.Write([]byte("Del success!"))
			}
		}

	case "Get":
/*
server side:
<- Get\n
<- ID\n
<- Filename\n
	->rawfile | ERROR

*/
		bufid,_,err:=rd.ReadLine()
		if err!=nil{
			fmt.Println("Get file: read id error --",err)
			return
		}
		var nID int64
		fmt.Sscanf(string(bufid),"%d",&nID)
		if buffile,_,err:=rd.ReadLine();err!=nil{
			fmt.Println("Get file: read filename error --",err)
		}else{
			strs := strings.Split(string(buffile), "../")
			path:= strs[len(strs)-1]	// for filesystem security
			dstfile,size,err:=backend.GetProjFile(nID,path)
			if err!=nil{
				conn.Write([]byte("ERROR :"+err.Error()))
			}else{
	//			conn.Write([]byte("OK\n"))
				srcfile,_:=os.Open(dstfile)
				io.CopyN(conn,srcfile,size)
				srcfile.Close()
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
				conn.Write([]byte("ERROR Browse proj error:"+err.Error()))
			}else{
				for _,line:=range(files){
					conn.Write([]byte(line+"\n"))
				}
			}
		}

	case "Update":
/*
server side:
<- Update\n
<- ID \n
<- projfile \n
	-> nOrgFileSize\n | ERROR No such file\n
<- nFileSize \n | CANCEL \n
<- RawFile
	-> OK | ERROR
*/
		bufid,_,errb:=rd.ReadLine()
		pfile,_,errp:=rd.ReadLine()
		if errb!=nil || errp!=nil {
			fmt.Println("Update proj: read update info error.")
			return
		}
		var nID int64
		fmt.Sscanf(string(bufid),"%d",&nID)
		if svrfile,size,err:=backend.GetProjFile(nID, string(pfile));err!=nil{
			conn.Write([]byte("ERROR cant access file--"+err.Error()))
		}else{
			conn.Write([]byte(fmt.Sprintf("%d\n",size)))
			line,_,err:=rd.ReadLine()
			if err!=nil{
				fmt.Println("Read file size error")
				return
			}
			sizebuf:=string(line)
			if sizebuf=="CANCEL"{
				return
			}
			var realsize int64
			if _,err:=fmt.Sscanf(sizebuf,"%d",&realsize);err!=nil{
				fmt.Println("Bad response, get file size error!")
			}else{
				tmpname:=fmt.Sprintf("/tmp/srcfile-%d",time.Now().UnixNano())
				tmpfile,_:=os.Create(tmpname)
				cpsize,err:=io.CopyN(tmpfile,rd,realsize)
				tmpfile.Close()
				defer os.Remove(tmpname) // may fail after rename, but doesn't matter if err!=nil || cpsize!=realsize {
				if err!=nil || cpsize!=realsize {
					fmt.Println("Update -- Copy file error:",err)
					conn.Write([]byte("ERROR copy file error"))
				}else{
					orginfo,_:=os.Stat(svrfile)
					if err:=os.Rename(tmpname,svrfile);err!=nil{
						fmt.Println("Write project source file error:",err)
						conn.Write([]byte("ERROR write source file error"))
					}else{
						os.Chmod(svrfile,orginfo.Mode())
						conn.Write([]byte("OK\n"))
					}
				}
			}
		}


	case "Edit":
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


		if bufid, _, err := rd.ReadLine(); err != nil {
			fmt.Println("Edit proj:read id error", err)
		} else {
			var nID int64
			fmt.Sscanf(string(bufid), "%d", &nID)
			if proj, err := backend.LookforID(nID); err != nil {
				conn.Write([]byte("ERROR Can't find id in db:" + err.Error()))
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
							conn.Write([]byte("ERROR Update database failed:" + err.Error()))
						} else {
							conn.Write([]byte("Edit success!"))
						}
					}
				} // else remote finish connection
			}
		}

	case "Search":
		if bufargs,_,err:=rd.ReadLine();err!=nil{
			fmt.Println("Search: read arg numbers error:",err)
		}else{
			nArgs :=0
			fmt.Sscanf(string(bufargs),"%d",&nArgs)
			keywords:=make([]string, 0, 20)
			for i:=0;i<nArgs;i++{
				if line,_,err:=rd.ReadLine();err!=nil{
					fmt.Println("Read args error:",err)
					return
				}else{
					keywords=append(keywords,string(line))
				}
			}

			projs, err := backend.SearchProj(keywords)
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
		}

	case "Run":
		//bufid:=make([]byte,10,10)
		if bufid, _, err := rd.ReadLine(); err != nil {
			fmt.Println("Run proj:read id error", err)
			//			conn.Write([]byte("ERROR "+err.Error())) // need not,since read error already
		} else {
			var nID int64
			var nParam int
			fmt.Sscanf(string(bufid), "%d%d", &nID, &nParam)
			params := make([]string, 0, 20)
			for i := 0; i < nParam; i++ {
				param, _, err := rd.ReadLine()
				if err != nil {
					fmt.Println("Get parameters failed.")
					conn.Write([]byte("ERROR Parameters error"))
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
		fmt.Println("Unknown command:", string(command))
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
