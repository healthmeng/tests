package main

import(
//"fmt"
_"html/template"
"io"
"net/http"
"log"
"fmt"
"backend"
)

func ListProjs(w io.Writer){
	fmt.Fprintf(w,"<!DOCTYPE html>\n<head>\n<title> Project list </title>\n<meta charset=\"utf-8\" />\n<body>\n<table id=\"projtbl\" border=\"1\">\n")
	projs,err:=backend.ListProj()

	if err!=nil{
		fmt.Fprintf(w,"%s\n",err.Error())
	}else{
		fmt.Fprintf(w,"<tr>\n<th>编号</th>\n<th>名称</th>\n<th>描述</th>\n<th>结论</th>\n<th>修改时间</th>\n</tr>\n")
		for _,proj :=range projs{
			fmt.Fprintf(w,"<tr>\n<td>%d</td>\n<td>%s</td>\n<td>%s</td>\n<td>%s</td>\n<td>%s</td>\n</tr>",
				proj.Id,proj.Title,proj.Descr,proj.Conclude,proj.Atime)
		}
	}
	fmt.Fprintf(w,"</table>\n</body>\n</head>\n")

}

func listproj(w http.ResponseWriter, r *http.Request){
	if r.Method=="GET"{
	//	t,_:=template.ParseFiles("list.htm")
//		tpl:=make(map[string] string)
	//	t.Execute(w,nil)
		ListProjs(w)
	}else{
	}
}

func main(){
	http.HandleFunc("/",listproj)
	http.HandleFunc("/list",listproj)
	if err:=http.ListenAndServe(":7777",nil);err!=nil{
		log.Println("Error:",err)
	}
}
