/* process object and database related operations. */

package backend

import (
_"github.com/Go-SQL-Driver/MySQL"
"database/sql"
"time"
"os"
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

func LookforID(id int64)(*PROJINFO, error){
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
        if err:=res.Scan(&proj.Id,&proj.Title,&proj.Descr,&proj.Atime,&proj.Conclude,&proj.Size,&proj.Path);err!=nil{
			fmt.Println("Scan error")
            return nil,err
		}
	}else{
			return nil,errors.New("Can't find record in db")
	}
	return proj,nil
}

func DelProj(id int64) error{
	db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests") //?charset=utf8")
    if err!=nil{
        fmt.Println("Open database failed")
        return err
    }
    defer db.Close()
	delcmd:=fmt.Sprintf("delete from proj where proj_id=%d",id)
	if result,err:=db.Exec(delcmd);err!=nil{
		fmt.Println("Exec delete cmd in db error:",err)
		return err
	}else{
		if rows,_:=result.RowsAffected();rows>0{
			projdir:=fmt.Sprintf("/opt/testssvr/%d",id)
			os.RemoveAll(projdir)
			return nil
		}
		return errors.New(fmt.Sprintf("Can't find project whose id=%d",id))
	}
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

func (info* PROJINFO)UpdateDB() error{
	db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests") //?charset=utf8")
    if err!=nil{
        fmt.Println("Open database failed")
        return err
    }
    defer db.Close()
	tm:=time.Now().Local()
	info.Atime=fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",tm.Year(),tm.Month(),tm.Day(),tm.Hour(),tm.Minute(),tm.Second())
	updatecmd:=fmt.Sprintf("update proj set title='%s',descr='%s',conclude='%s',projtime='%s' where proj_id=%d",info.Title,info.Descr,info.Conclude,info.Atime,info.Id)
	if result,err:=db.Exec(updatecmd);err!=nil{
		fmt.Println("Exec update cmd in db error:",err)
		return err
	}else{
		if rows,_:=result.RowsAffected();rows>0{
			return nil
		}else{
			return errors.New("Can't find project in database")
		}
	}
}

func (info* PROJINFO) CreateInDB() error{
	db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests")

	//db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests?charset=utf8")  // this will get messed code in Chinese
	if err!=nil{
		fmt.Println("Open database failed")
		return err
	}
	defer db.Close()
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
