package backend 

import (
"fmt"
"time"
)


type PROJINFO struct{
	Id int
	Title string
	Atime string // always use database updatetime
	Descr string
	Conclude string
	Path string
	IsDir bool
}


func (info* PROJINFO) CreateInDB() error{
 info.Id=00010
 info.Atime=time.Now().String()
 fmt.Println("create")
 return nil
}
