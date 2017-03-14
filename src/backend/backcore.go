/* process object and database related operations. */

package backend

import (
	"database/sql"
	"log"
	"errors"
	"fmt"
	_ "github.com/Go-SQL-Driver/MySQL"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//var dblogin string = "work:abcd1234@tcp(localhost:3306)/tests"
//var dblogin string = "work:abcd1234@tcp(123.206.55.31:3306)/tests"
var hdb *sql.DB =nil

type PROJINFO struct {
	Id       int64
	Title    string
	Atime    string // always use database updatetime
	Descr    string
	Conclude string
	Path     string
	IsDir    bool
	Size     int64
}

func getDB()(*sql.DB, error){
	dbdrv := "mysql"
	dblogin := "work:abcd1234@tcp(localhost:3306)/tests"
	var err error=nil
	if hdb==nil{
		hdb, err = sql.Open(dbdrv, dblogin)
		if err != nil {
			log.Println("Open database failed")
			hdb=nil
		}
	}
	return hdb,err
}

func getProjDir(id int64) string {
	return fmt.Sprintf("/opt/testssvr/%d", id)
}

func (info *PROJINFO) getRootDir() string {
	basedir := getProjDir(info.Id)
	rootdir := basedir + "/" + info.Path
	path := rootdir
	for {
		sp := strings.Replace(path, "//", "/", -1)
		if sp != path {
			path = sp
		} else {
			break
		}
	}
	rootdir = strings.TrimSuffix(path, "/")
	return rootdir
}

func GetProjFile(id int64, file string) (string, int64, error) {
	proj, err := LookforID(id)
	if err != nil {
		return "", 0, err
	}
	return proj.GetDestFile(file)
}

func (info *PROJINFO) GetDestFile(relapath string) (string, int64, error) {
	//	destfile=rootprefix+relapath
	rootdir := info.getRootDir()
	if finfo, err := os.Stat(rootdir); err != nil {
		return "", 0, err
	} else {
		destFile := strings.TrimSuffix(rootdir, finfo.Name()) + relapath
		dinfo, err := os.Stat(destFile)
		if err != nil {
			return "", 0, err
		} else if dinfo.IsDir() {
			return "", 0, errors.New("Dir transfer is not supported")
		} else {
			return destFile, dinfo.Size(), nil
		}
	}
}

func (info *PROJINFO) diskFiles() (string, []string, error) {
	rootdir := info.getRootDir()

	output, err := exec.Command("find", rootdir).Output()
	if err != nil {
		log.Println("browse root dir error:", err)
		return rootdir, nil, err
	}
	lines := strings.Split(string(output), "\n")
	return rootdir, lines, nil
}

func makeOutput(rootdir string, dfiles []string) ([]string, error) {
	if finfo, err := os.Stat(rootdir); err != nil {
		log.Println(rootdir, " not exist")
		return nil, err
	} else {
		dirname := finfo.Name()
		result := make([]string, 0, 50)
		for _, path := range dfiles {
			if path == "" {
				continue
			}
			pinfo, err := os.Stat(path)
			if err != nil {
				continue
			}
			isdir := pinfo.IsDir()
			relatedir := strings.TrimPrefix(path, rootdir)
			final := dirname + relatedir
			if isdir {
				result = append(result, "[dir] "+final)
			} else {
				result = append(result, fmt.Sprintf("      [%d|%o] %s", pinfo.Size(), pinfo.Mode(), final))
			}
		}
		return result, nil
	}
}

func BrowseProj(id int64) ([]string, error) {
	proj, err := LookforID(id)
	if err != nil {
		log.Println("Lookfor proj failed:", err)
		return nil, err
	}
	if rootdir, dfiles, err := proj.diskFiles(); err != nil {
		log.Println("Check disk file error:", err)
		return nil, err
	} else {
		if output, err := makeOutput(rootdir, dfiles); err != nil {
			return nil, err
		} else {
			return output, nil
		}
	}
}

func LookforID(id int64) (*PROJINFO, error) {
	db, err :=getDB()
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("select * from proj where proj_id=%d", id)
	proj := new(PROJINFO)
	res, _ := db.Query(query)
	defer res.Close()
	if res.Next() {
		if err := res.Scan(&proj.Id, &proj.Title, &proj.Descr, &proj.Atime, &proj.Conclude, &proj.Path); err != nil {
			//if err := res.Scan(&proj.Id, &proj.Title, &proj.Descr, &proj.Atime, &proj.Conclude, &proj.Size, &proj.Path); err != nil {
			log.Println("Scan error")
			return nil, err
		}
	} else {
		return nil, errors.New("Can't find record in db")
	}
	return proj, nil
}

func DelProj(id int64) error {
	db, err :=getDB()
	if err != nil {
		return err
	}
	delcmd := fmt.Sprintf("delete from proj where proj_id=%d", id)
	if result, err := db.Exec(delcmd); err != nil {
		log.Println("Exec delete cmd in db error:", err)
		return err
	} else {
		if rows, _ := result.RowsAffected(); rows > 0 {
			projdir := getProjDir(id)
			os.RemoveAll(projdir)
			return nil
		}
		return errors.New(fmt.Sprintf("Can't find project whose id=%d", id))
	}
}

func SearchProj(keywords []string, cs bool) ([]PROJINFO, error) {
	db, err :=getDB()
	//db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests?charset=utf8")
	if err != nil {
		return nil, err
	}

	query := "select * from proj where "
	for index, arg := range keywords {
		if index != 0 {
			query += " AND "
		}
		query += fmt.Sprintf("(title like '%%%s%%' OR descr like '%%%s%%' OR projtime like '%%%s%%' OR conclude like '%%%s%%' OR path like '%%%s%%' ", arg, arg, arg, arg, arg)
		_, err := strconv.Atoi(arg)
		if err == nil {
			query += (" OR proj_id = " + arg)
		}
		query += ")"
	}
	if cs{
		query+=" collate utf8_bin"
	}
	res, _ := db.Query(query)
	defer res.Close()

	projs := make([]PROJINFO, 0, 100)
	for i := 0; res.Next(); i++ {
		proj := new(PROJINFO)
		proj.IsDir = true
		err = res.Scan(&proj.Id, &proj.Title, &proj.Descr, &proj.Atime, &proj.Conclude, &proj.Path)
		if err != nil {
			return nil, err
		}
		projs = append(projs, *proj)
	}
	return projs, nil
}

func ListProj() ([]PROJINFO, error) {
	db, err :=getDB()
	if err != nil {
		return nil, err
	}
	query := "select count(*) as value from proj"
	var rows int64
	if err := db.QueryRow(query).Scan(&rows); err != nil {
		log.Println("Query rows error")
		return nil, err
	}
	query = "select * from proj"
	projs := make([]PROJINFO, rows, rows)
	res, _ := db.Query(query)
	defer res.Close()
	for i := 0; res.Next(); i++ {
		projs[i].IsDir = true
		err = res.Scan(&projs[i].Id, &projs[i].Title, &projs[i].Descr, &projs[i].Atime, &projs[i].Conclude, &projs[i].Path)
		if err != nil {
			return nil, err
		}
	}
	return projs, nil
}

func (info *PROJINFO) InitDir(tmpfile string) error {
	projdir := getProjDir(info.Id)
	exec.Command("mkdir", "-p", projdir).Run()
	return exec.Command("tar", "xzvf", tmpfile, "-C", projdir).Run()
}

func (info *PROJINFO) UpdateDB() error {
	db, err :=getDB()
	if err != nil {
		return err
	}
	tm := time.Now().Local()
	info.Atime = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
	updatecmd := fmt.Sprintf("update proj set title='%s',descr='%s',conclude='%s',projtime='%s' where proj_id=%d", info.Title, info.Descr, info.Conclude, info.Atime, info.Id)

	updatecmd=strings.Replace(updatecmd,"\\","\\\\",-1)
	if result, err := db.Exec(updatecmd); err != nil {
		log.Println("Exec update cmd in db error:", err)
		return err
	} else {
		if rows, _ := result.RowsAffected(); rows > 0 {
			return nil
		} else {
			return errors.New("Can't find project in database")
		}
	}
}

func (info *PROJINFO) CreateInDB() error {
	db, err := getDB()
	//db,err:=sql.Open("mysql","work:abcd1234@tcp(123.206.55.31:3306)/tests?charset=utf8")  // this will get messed code in Chinese
	if err != nil {
		return err
	}
	tm := time.Now().Local()
	info.Atime = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
	//	st,err:=db.Prepare("insert proj set title=?,descr=?,projtime=?,conclude=?,
	query := fmt.Sprintf("insert into proj (title,descr,projtime,conclude,path) values ('%s','%s','%s','%s','%s')", info.Title, info.Descr, info.Atime, info.Conclude, info.Path)
	query=strings.Replace(query,"\\","\\\\",-1)
	if result, err := db.Exec(query); err == nil {
		info.Id, _ = result.LastInsertId()
		return nil
	} else {
		log.Println("insert failed")
		return err
	}
}
