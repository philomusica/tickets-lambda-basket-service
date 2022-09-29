.PHONY: deps clean build deploy test vet fmt
BINDIR:=./bin
ZIPFILE:=function.zip
BINARY:=main

deps:
	go get -u ./...

clean:
	rm -rf $(BINDIR)

build:
	GOOS=linux GOARCH=amd64 go build -o $(BINDIR)/$(BINARY) ./...

deploy: build
ifeq ($(ARN),)
	@echo "Please set the ARN"
else
	zip -r $(ZIPFILE) $(BINDIR)
	aws lambda update-function-code --function-name $(ARN) --zip-file fileb://$(ZIPFILE)
endif

test:
	go test -v -coverprofile cover.out ./...

vet:
	go vet ./...

fmt:
	go fmt ./...