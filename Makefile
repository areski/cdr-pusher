BINARY = ./bin/cdr-pusher

install-daemon:
	go install .

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
	cp -i cdr-pusher.yaml /etc/cdr-pusher.yaml

logdir:
	@mkdir /var/log/cdr-pusher

get:
	@go get -d .

build: get configfile
	@mkdir -p bin
	@go build -a -o bin/cdr-pusher .
