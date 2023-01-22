.PHONY: deps clean build deploy test vet fmt
BINDIR:=./bin
ZIPFILE:=function.zip
BINARY:=main
REPORT:=./report

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
	go test -v -cover ./...

cover:
	mkdir -p $(REPORT)
	go test ./... -coverprofile $(REPORT)/cover.out
	go tool cover -html=$(REPORT)/cover.out -o $(REPORT)/index.html
	cd $(REPORT) && python3 -m http.server 8000
	
vet:
	go vet ./...

fmt:
	go fmt ./...
