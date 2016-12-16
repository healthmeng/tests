/* process object and database related operations. */

package backend

import (
_"github.com/Go-SQL-Driver/MySQL"
"database/sql"
"time"
"os"
"os/exec"
"strings"
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

func getProjDir(id int64) string{
	return fmt.Sprintf("/opt/testssvr/%d",id)
}

func BrowseProj(id int64)([]string,error){
	proj,err:=LookforID(id)
	if err!=nil{
		fmt.Println("Lookfor proj failed:",err)
		return nil,err
	}
	basedir:=getProjDir(id)
	rootdir:=basedir+"/"+proj.Path
    path := rootdir
    for {
        sp := strings.Replace(path, "//", "/", -1)
        if sp != path {
            path = sp
        } else {
            break
        }
    }
	rootdir=strings.TrimSuffix(path,"/")

	if info,err:=os.Stat(rootdir);err!=nil{
		fmt.Println(rootdir," not exist")
		return nil,err
	}else{
		dirname:=info.Name()
		output,err:=exec.Command("find",rootdir).Output()
		if err!=nil{
			fmt.Println("browse root dir error:",err)
			return nil,err
		}
		lines:=strings.Split(string(output),"\n")
		result:=make([]string,0,50)
		for _,path:=range(lines){
			if path==""{
				continue
			}
			pinfo,err:=os.Stat(path)
			if err!=nil{
				continue
			}
			isdir:=pinfo.IsDir()
			reldir:=strings.TrimPrefix(path,rootdir)
			final:=dirname+reldir
			if isdir{
				result=append(result,"[dir]"+final)
			}else{
				result=append(result,final)
			}
		}
		return result,nil
	}
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
			projdir:=getProjDir(id)
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

func (info* PROJINFO)InitDir(tmpfile string) error{
	projdir:=getProjDir(info.Id)
	exec.Command("mkdir", "-p", projdir).Run()
	return exec.Command("tar","xzvf",tmpfile,"-C",projdir).Run()
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
