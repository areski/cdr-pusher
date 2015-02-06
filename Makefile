BINARY = ./bin/cdr-pusher

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

configfile:
	cp cdr-pusher.yaml /etc/cdr-pusher.yaml

get:
	@go get -d ./...

build: get
	@mkdir -p bin
	@go build -a -o bin/cdr-pusher .