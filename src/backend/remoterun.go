package backend

import (
"fmt"
"errors"
"time"
"io"
"os"
"os/exec"
)


type Redirect interface{
	GetInput(inpipe io.WriteCloser)// get remote input
	SendOutput(outpipe io.ReadCloser) // get local output
}

func waitOut(chok chan int,cmd *exec.Cmd){
	cmd.Wait()
	chok<-0 // inform timeout cleaner
}

func RunID(id int64,rio Redirect)(chan int ,error){
	chout:=make(chan int,1)
	chok:=make(chan int,1)
//	cmd:=exec.Command("/tmp/deploy")
	proj,err:=lookforID(id)
	if err!=nil{
		fmt.Println("Can't find project-- id:",id)
		return chout,err
	}
	cid,cmd,err:=proj.createContainer() // create container,copy file, return container (id string), and prepare run
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
	select{
		case <-chok:
			// nothing, process quit normally
		case <-time.After(time.Second*60):
			// time out, kill container
	}
	// delete cid container
		if err:=removeContainer(cid);err!=nil{
			fmt.Println("Remove container error:",err)
		}
	// inform server, process over
		chout<-1
}


func (proj* PROJINFO)Uncompress()(string,error){
	basePath:=fmt.Sprintf("/opt/testssvr/%d/",proj.Id)
	srcPath:=basePath+"proj.tgz"
	if _,err:=os.Stat(srcPath);err!=nil{
		return "",errors.New("Can't find "+srcPath)
	}
	cmd:=exec.Command(fmt.Sprintf("tar xzvf %s -C %s",srcPath,basePath))
	if err:=cmd.Run();err!=nil{
		fmt.Println("tar proj.tgz error:",err)
		return "",err
	}
	dstPath:=basePath+proj.Path
	finfo,err:=os.Stat(dstPath)
	if err!=nil{
		fmt.Println(dstPath, "not found, check your tar parameters.")
		return dstPath,err
	}
	proj.IsDir=finfo.IsDir()
	return dstPath,nil
}

func removeContainer(cid string) error{
	cmd:=exec.Command("docker rm "+cid)
	return cmd.Run()
}

func (proj* PROJINFO)createContainer()(string,*exec.Cmd,error){
	obsPath,err:=proj.Uncompress()  // uncompress, stat(isdir),return obsolute path
	if(err!=nil){
		fmt.Println("Uncompress proj.tgz error",err)
		return "",nil,err
	}
	defer os.RemoveAll(obsPath)
	ctrun:=""
	ctwork:=""
	if proj.IsDir{
		if finfo,err:=os.Stat(obsPath+"/run");err==nil{	// host path
			if !finfo.IsDir() && (finfo.Mode() &0700 !=0){
				ctrun="/tmp/"+proj.Path+"/run"	// guest path
				ctwork="/tmp/"+proj.Path
			}
		}else{
			return "",nil,errors.New("File 'run' not found in your directory.")
		}
	}else{// get postfix, and try to build them, then copy runable binary to container
		return  "", nil, errors.New("Single file will support later")
	}
	if ctrun==""{
		return "",nil,errors.New("Can not run because of incorrect perm or directory structure")
	}
	// docker create -i -t devel: get id
	crcmd:=exec.Command("docker create -i -t devel -w "+ctwork+" "+ctrun) // add run command, workdir
	outbyte,err:=crcmd.Output()
	ctid:=string(outbyte)
	if err!=nil{
		fmt.Println("Create contailer error:",err)
		return "",nil,err
	}
	crcmd=exec.Command("docker cp "+obsPath+" "+ctid+":/tmp")
	if err:=crcmd.Run();err!=nil{
		fmt.Println("docker cp error:",err)
		return "",nil,err
	}
	crcmd=exec.Command("docker start -a -i "+ctid)
	return ctid,crcmd,nil
}

