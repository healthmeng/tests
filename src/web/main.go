package main

import(
//"fmt"
"html/template"
"net/http"
"log"
)

func listproj(w http.ResponseWriter, r *http.Request){
	if r.Method=="GET"{
		t,_:=template.ParseFiles("list.htm")
//		tpl:=make(map[string] string)
		t.Execute(w,nil)
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
