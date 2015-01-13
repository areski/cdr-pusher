BINARY = ./bin/fs-pusher

install-daemon:
	go install ./...

deps:
	go get .

clean:
	rm $(BINARY)

test:
	go test .
	golint

servedoc:
	godoc -http=:6060

get:
	@go get -d ./...

build: get
	@mkdir -p bin
	@go build -a -o bin/fs-pusher .