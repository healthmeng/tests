set GOPATH=%~dp0
go get github.com/Go-SQL-Driver/MySQL 
#cd %~dp0/bin
go build tests
go build server
