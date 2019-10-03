# Go parameters
GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
GOGET=go get
BINARY_NAME=rpicar

all: test build

build:
	env GOOS=linux GOARCH=arm GOARM=5 CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 $(GOBUILD) -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ../
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
publish:
	scp $(BINARY_NAME) pi@192.168.178.23:/home/pi/
