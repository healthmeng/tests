package backend 

import (
_"github.com/Go-SQL-Driver/MySQL"
"database/sql"
"fmt"
"time"
"io"
"os/exec"
)


type PROJINFO struct{
	Id int64
	Title string
	Atime string // always use database updatetime
	Descr string
	Conclude string
	Path string
	IsDir bool
	Size int64
}

type Redirect interface{
	GetInput(inpipe io.WriteCloser)// get remote input
	SendOutput(outpipe io.ReadCloser) // get local output
}

func waitOut(chok chan int,cmd *exec.Cmd){
	cmd.Wait()
	chok<-0
}

// todo: lookforID, createImage,

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

	// inform server process over
		chout<-1
}

func ListProj()([]PROJINFO,error){
    db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests") //?charset=utf8")
    //db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests?charset=utf8")
    if err!=nil{
        fmt.Println("Open database failed")
        return nil,err
    }
    defer db.Close()
    query:="select count(*) as value from proj" ;
	var rows int64
    if err:=db.QueryRow(query).Scan(&rows);err!=nil{
		fmt.Println("Query rows error")
		return nil,err
	}
	query="select * from proj"
	projs:=make([]PROJINFO,rows,rows)
	res,_:=db.Query(query)
	defer res.Close()
	for i:=0;res.Next();i++{
		projs[i].IsDir=true
		err=res.Scan(&projs[i].Id,&projs[i].Title,&projs[i].Descr,&projs[i].Atime,&projs[i].Conclude,&projs[i].Size,&projs[i].Path)
		if err!=nil{
			return nil,err
		}
	}
	return projs,nil
}

func (info* PROJINFO) CreateInDB() error{
	db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests") 
	//db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests?charset=utf8")  // this will get messed code in Chinese
	if err!=nil{
		fmt.Println("Open database failed")
		return err
	}
	defer db.Close()
	fmt.Println("Title=",info.Title)
	tm:=time.Now().Local()
	info.Atime=fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",tm.Year(),tm.Month(),tm.Day(),tm.Hour(),tm.Minute(),tm.Second())
//	st,err:=db.Prepare("insert proj set title=?,descr=?,projtime=?,conclude=?,
	query:=fmt.Sprintf("insert into proj (title,descr,projtime,conclude,fsize,path) values ('%s','%s','%s','%s',%d,'%s')",info.Title,info.Descr,info.Atime,info.Conclude,info.Size,info.Path)
	if result,err:=db.Exec(query);err==nil{
		info.Id,_=result.LastInsertId()
		return nil
	}else{
	fmt.Println("insert failed")
		return err
	}
}
