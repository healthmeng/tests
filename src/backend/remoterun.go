// process remote run process in container and io redirection related operations
package backend

import (
"fmt"
"strings"
"errors"
"time"
"io"
"os"
"os/exec"
"backend/runsrc"
_"backend/rungo"
_"backend/runcncpp"
_"backend/runpy"
_"backend/runscript"
)


type Redirect interface{
	GetInput(inpipe io.WriteCloser)// get remote input
	SendOutput(outpipe io.ReadCloser) // get local output
}

func waitOut(chok chan int,cmd *exec.Cmd){
	cmd.Wait()
	chok<-0 // inform timeout cleaner
}

func GetPlugins() []string{
	return runsrc.GetSupport()
}

func RunID(id int64,rio Redirect,params []string)(chan int ,error){
	chout:=make(chan int,1)
	chok:=make(chan int,1)
	proj,err:=LookforID(id)
	if err!=nil{
		fmt.Println("Can't find project-- id:",id)
		return chout,err
	}
	cid,cmd,err:=proj.createContainer(params) // create container,copy file, return container (id string), and prepare run
	if err!=nil{
		fmt.Println("Prepare run command error:",err)
		return  chout,err
	}
	outp,_:=cmd.StdoutPipe()
	inp,_:=cmd.StdinPipe()
	go rio.SendOutput(outp)
	go rio.GetInput(inp)
	if err:=cmd.Start();err!=nil{
	  return chout,err
	}
	go waitOut(chok,cmd)
	go procTimeout(cid,chok,chout)
	return chout,nil
}

func procTimeout(cid string,chok , chout chan int){
	// sleep xx second and kill container
	normal:=true
	select{
		case <-chok:
			// nothing, process quit normally
		case <-time.After(time.Second*30):
			normal=false;
			// time out, kill container
	}
	// delete cid container
		if err:=removeContainer(cid,normal);err!=nil{
			fmt.Println("Remove container error:",err)
		}
	// inform server, process over
		chout<-1
}


func (proj* PROJINFO)prepareFile()(string,string,error){// abs path & final name
	basePath:=getProjDir(proj.Id)+"/"
	dstPath:=basePath+proj.Path
	finfo,err:=os.Stat(dstPath)
	if err!=nil{
		fmt.Println(dstPath, "Running file(s) not found.")
		return dstPath,"",err
	}
	proj.IsDir=finfo.IsDir()
	return dstPath,finfo.Name(),nil
}

func removeContainer(cid string, exited bool) error{
	if !exited{
		cmd:=exec.Command("docker","kill",cid)
		cmd.Run()
	}
	// if kill failed, still need to try  rm
	cmd:=exec.Command("docker", "rm",cid)
	return cmd.Run()
}

func (proj* PROJINFO)createContainer(args []string)(string,*exec.Cmd,error){
	obsPath,fName,err:=proj.prepareFile()  // uncompress, stat(isdir),return obsolute path
	if(err!=nil){
		fmt.Println("Prepare running file(s) error",err)
		return "",nil,err
	}
	ctrun:=""
	ctwork:="/tmp/"
	strcmdfile:=getProjDir(proj.Id)+"/run"
	if proj.IsDir{
		if finfo,err:=os.Stat(obsPath+"/run");err==nil{	// host path
			if !finfo.IsDir() && (finfo.Mode() &0700 !=0){
				ctrun="/tmp/"+fName+"/run"	// guest path
				ctwork+=fName
			}
		}else{
			return "",nil,errors.New("File 'run' not found in your directory.")
		}
	}else{// get postfix, and try to build them, then copy runable binary to container
		srccmd:=runsrc.GetCmd(obsPath,args...)
		cmdfile,_:=os.Create(strcmdfile)
		cmdfile.Write([]byte(srccmd))
		cmdfile.Close()
		os.Chmod(strcmdfile,0777)
		defer os.Remove(strcmdfile)
		ctrun="/tmp/run"
	}
	if ctrun==""{
		return "",nil,errors.New("Can not run because of incorrect perm or directory structure")
	}
	// docker create -i -w  workdir devel /bin/bash -c cmd ,return : get vmid
		crcmd:=exec.Command("docker","create","-i","-w",ctwork,"devel","/bin/bash","-c",ctrun) //if add "-t",input message will loop to output again
		//crcmd:=exec.Command("docker","create","-i","-t","-w",ctwork,"devel","/bin/bash","-c",ctrun) // add run command, workdir
	outbyte,err:=crcmd.Output()
	if err!=nil{
		fmt.Println("Create container error:",err)
		return "",nil,err
	}
	ctid:=strings.Replace(string(outbyte),"\n","",-1)
	ctid=strings.Replace(ctid,"\r","",-1)
	// copy source
	crcmd=exec.Command("docker","cp",obsPath,ctid+":/tmp")
	if err:=crcmd.Run();err!=nil{
		fmt.Println("docker cp src error:",err)
		return "",nil,err
	}
	// create run command for src
	if !proj.IsDir{
		crcmd=exec.Command("docker","cp",strcmdfile,ctid+":/tmp")
		if err:=crcmd.Run();err!=nil{
			fmt.Println("docker cp cmdfile error:",err)
			return "",nil,err
		}
	}

	crcmd=exec.Command("docker","start", "-i",ctid)
	return ctid,crcmd,nil
}

