package main

import(
//"fmt"
_"html/template"
"strings"
"io"
"net/http"
"os"
"log"
"fmt"
"backend"
)

var dhandle http.Handler

func ListProjs(w io.Writer){
	fmt.Fprintf(w,"<!DOCTYPE html>\n<head>\n<title> Project list </title>\n<meta charset=\"utf-8\" />\n<body>\n<table id=\"projtbl\" border=\"1\">\n")
	fmt.Fprintf(w,"<form name=\"lists\" method=\"post\" action=\"detail\">\n")
	fmt.Fprintf(w,"<input type=\"hidden\" name=\"sel\" id=\"projid\" />\n")
	projs,err:=backend.ListProj()

	if err!=nil{
		fmt.Fprintf(w,"%s\n",err.Error())
	}else{
		fmt.Fprintf(w,"<tr>\n<th>编号</th>\n<th>名称</th>\n<th>描述</th>\n<th>结论</th>\n<th>修改时间</th>\n<th>运行</th></tr>\n")
		for _,proj :=range projs{
			fmt.Fprintf(w,"<tr>\n<td><a href=\"javascript:lists.sel.value=%d;lists.submit()\">%d</a></td>\n<td>%s</td>\n<td>%s</td>\n<td>%s</td>\n<td>%s</td>\n<td><a href=\"javascript:lists.sel.value='run:%d';lists.submit()\"><img src=\"http://uus-img6.android.d.cn/content_pic/201605/behpic/icon/420/2-71420/icon-1463728370333.png\" height=\"24\" width=\"24\" /></a></td></tr>",
				proj.Id,proj.Id,proj.Title,proj.Descr,proj.Conclude,proj.Atime,proj.Id)
		}
	}
	fmt.Fprintf(w,"</form>\n</table>\n</body>\n</head>\n")

}

func listproj(w http.ResponseWriter, r *http.Request){
	if r.Method=="GET"{
	//	t,_:=template.ParseFiles("list.htm")
//		tpl:=make(map[string] string)
	//	t.Execute(w,nil)
		ListProjs(w)
	}else{
		r.ParseForm()
		obj:=r.Form["sel"][0]
		if strings.HasPrefix(obj,"run:"){
			obj=strings.TrimPrefix(obj,"run:")
			fmt.Fprintf(w,"selected : %s\n",obj)
		}else{
			http.Redirect(w,r,fmt.Sprintf("/projs/%s",obj),http.StatusFound)
		}
	}
}

func InitDB(){
	config,err:=os.Open("db.ini")
	if err==nil{
		defer config.Close()
		var host,user,passwd string
		fmt.Fscanf(config,"%s%s%s",&host,&user,&passwd)
		if host!="" && user !="" && passwd!=""{
			backend.ChangeDefDB(host,user,passwd)
		}
	}
}

func browser(w http.ResponseWriter, r *http.Request){
	dhandle.ServeHTTP(w,r)
}

func main(){
	InitDB()
	http.HandleFunc("/",listproj)
	http.HandleFunc("/list",listproj)
	dhandle=http.StripPrefix("/projs",http.FileServer(http.Dir("/opt/testssvr")))
	http.HandleFunc("/projs/",browser)
//	http.HandleFunc("/detail",showdetail)
	if err:=http.ListenAndServe(":7777",nil);err!=nil{
		log.Println("Error:",err)
	}
}
