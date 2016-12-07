package backend 

import (
_"github.com/Go-SQL-Driver/MySQL"
"database/sql"
"fmt"
"time"
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

func ListProj()([]PROJINFO,error){
	rows,err:=GetDBRows()
	if err!=nil{
		fmt.Println("Get database rows error")
		return nil,err
	}
	projs:=make([]PROJINFO,rows,rows)
	return projs,nil
}

func (info* PROJINFO) CreateInDB() error{
	db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests?charset=utf8")
	if err!=nil{
		fmt.Println("Open database failed")
		return err
	}
	defer db.Close()
	tm:=time.Now().Local()
	info.Atime=fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",tm.Year(),tm.Month(),tm.Day(),tm.Hour(),tm.Minute(),tm.Second())
//	st,err:=db.Prepare("insert proj set title=?,descr=?,projtime=?,conclude=?,
	query:=fmt.Sprintf("insert into proj (title,descr,projtime,conclude,fsize) values ('%s','%s','%s','%s',%d)",info.Title,info.Descr,info.Atime,info.Conclude,info.Size)
	if result,err:=db.Exec(query);err==nil{
		info.Id,_=result.LastInsertId()
		return nil
	}else{
	fmt.Println("insert failed")
		return err
	}
}
