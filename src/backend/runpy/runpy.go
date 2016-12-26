package runpy

import (
"fmt"
"os"
"bufio"
"strings"
"backend/runsrc"
)

func init(){
	runsrc.Register("Python",Judgement,RetCmd)
}

func Judgement(srcpath string) int{
	if strings.HasSuffix(srcpath,".py") ||
	 strings.HasSuffix(srcpath,".PY") {
		return 1
	}else{
        file,_:=os.Open(srcpath)
        defer file.Close()
        rd:=bufio.NewReader(file)
        for i:=0;i<10;i++{
            line,_,err:=rd.ReadLine()
            if err!=nil{
                break
            }
            if strings.HasPrefix(string(line),"#!/usr/bin/python"){
                return 1
            }
        }
    }
	return 0
}

func RetCmd(tp int,srcpath string,args ...string)string{
	st,_:=os.Stat(srcpath)
	ret:=fmt.Sprintf("/usr/bin/python /tmp/%s",st.Name())
	for _,arg:=range(args){
		ret+=" "+arg
	}
	return ret
}
