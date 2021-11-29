export GOPATH=$(PWD)

test_obj:tserver tests web

tserver:
	go build tserver 
tests:
	go build tests 

web:
	go build web
install:
	go install tserver tests web

clean:
	rm -f bin/*
	rm -f tserver tests web
