package runsrc

import (
//"fmt"
"errors"
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

func GetCmd(srcname string, args ...string)(string,error){
	tp:=0
	ret:=""
	for _,obj:=range(handler){
		if tp=obj.CanProc(srcname);tp!=0{
			ret=obj.GetRunString(tp,srcname,args...)
			break
		}
	}
	if tp==0{
		return "",errors.New("Can't find source code running policy")
	}
	return ret,nil
}
