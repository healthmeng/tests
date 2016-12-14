package runcncpp

import (
"fmt"
"os"
"strings"
"backend/runsrc"
)

func init(){
	runsrc.Register("C/C++",Judgement,RetCmd)
}

func Judgement(srcpath string) int {
	if strings.HasSuffix(srcpath,".c") || strings.HasSuffix(srcpath,".C"){
		return 1
	}
	if strings.HasSuffix(srcpath,".CPP") ||	 strings.HasSuffix(srcpath,".cpp") {
		return 2
	}else{
		return 0
	}
}

func RetCmd(tp int,srcpath string,args ...string)string{
	st,_:=os.Stat(srcpath)
	compiler:="/usr/bin/gcc"
	if tp==2{
		compiler="/usr/bin/g++"
	}
	ret:=fmt.Sprintf("%s /tmp/%s -o /tmp/prog\n/tmp/prog",compiler,st.Name())
	for _,arg:=range(args){
		ret+=" "+arg
	}
	return ret
}
