export GOPATH=`pwd`
go get github.com/Go-SQL-Driver/MySQL 
cd bin
go build tests
go build server
