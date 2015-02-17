
all:
	@GOPATH=`pwd` go install -v ./...

fmt:
	@GOPATH=`pwd` go fmt ./...

clean:
	@rm -rf bin pkg
