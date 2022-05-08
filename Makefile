
OUTPUT_NAME = go-transact

build:
	go build -o ${OUTPUT_NAME} -ldflags "-s -w"

install:
	go install

clean:
	go clean

clean-all:
	go clean --cache -testcache

# Tidy up dependencies
tidy:
	go mod tidy