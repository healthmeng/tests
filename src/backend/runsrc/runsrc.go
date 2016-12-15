// provide a plugin framework for different source code execution

package runsrc

import (
"fmt"
"os"
)

type SRCDETECT struct{
	Detector string
	CanProc func (srcname string) int 
	GetRunString func(tp int,srcname string,args ...string)(string)
}

var handler []SRCDETECT

func Register(detector string, match func(string) int ,proc func(tp int ,srcname string,args ...string) string){
	obj:=SRCDETECT{detector, match,proc}
	handler=append(handler,obj)
}

func init(){
	handler=make([]SRCDETECT,0,50)
}

func GetCmd(srcname string, args ...string)string{
	tp:=0
	var ret string
	info,_:=os.Stat(srcname)
	filename:=info.Name()
	for _,obj:=range(handler){
		if tp=obj.CanProc(srcname);tp!=0{
			ret=fmt.Sprintf("echo \"Trying to run %s as %s source code:\"\n",filename,obj.Detector)+obj.GetRunString(tp,srcname,args...)
			break
		}
	}
	if tp==0{
		ret=fmt.Sprintf("echo \"Don't know how to run %s\"",filename)
	}
	return ret
}
