package backend 

import (
_"github.com/Go-SQL-Driver/MySQL"
"database/sql"
"time"
"fmt"
"errors"
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

func lookforID(id int64)(*PROJINFO, error){
    db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests")
    if err!=nil{
        fmt.Println("Open database failed")
        return nil,err
    }
    defer db.Close()
    query:=fmt.Sprintf("select * from proj where proj_id=%d",id)
    proj:=new(PROJINFO)
    res,_:=db.Query(query)
    defer res.Close()
	if res.Next(){
        if err:=res.Scan(proj.Id,&proj.Title,&proj.Descr,&proj.Atime,&proj.Conclude,&proj.Size,&proj.Path);err!=nil{
            return nil,err
		}
	}else{
			return nil,errors.New("Can't find record in db")
	}
	return proj,nil
}

func ListProj()([]PROJINFO,error){
    db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests") //?charset=utf8")
    //db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests?charset=utf8")
    if err!=nil{
        fmt.Println("Open database failed")
        return nil,err
    }
    defer db.Close()
    query:="select count(*) as value from proj" 
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
