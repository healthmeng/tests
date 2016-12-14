package runscript

import (
"fmt"
"os"
"strings"
"bufio"
"backend/runsrc"
)

func init(){
	runsrc.Register("Shell/Perl",Judgement,RetCmd)
}

func Judgement(srcpath string) int{
	if strings.HasSuffix(srcpath,".pl") ||
	 strings.HasSuffix(srcpath,".PL") {
		return 2
	}else if strings.HasSuffix(srcpath,".sh") || strings.HasSuffix(srcpath,".SH"){
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
			if strings.HasPrefix(string(line),"#!/usr/bin/perl"){
				return 2
			}
		}
		return 0
	}
}

func RetCmd(tp int,srcpath string,args ...string)string{
	st,_:=os.Stat(srcpath)
	runner:="/bin/bash"
	if tp==2{
		runner="/usr/bin/perl"
	}
	ret:=fmt.Sprintf("%s /tmp/%s",runner,st.Name())
	for _,arg:=range(args){
		ret+=" "+arg
	}
	return ret
}
