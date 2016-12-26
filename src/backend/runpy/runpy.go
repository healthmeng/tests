package runpy

import (
"fmt"
"os"
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
		return 0
	}
}

func RetCmd(tp int,srcpath string,args ...string)string{
	st,_:=os.Stat(srcpath)
	ret:=fmt.Sprintf("/usr/bin/python /tmp/%s",st.Name())
	for _,arg:=range(args){
		ret+=" "+arg
	}
	return ret
}
