run: build
	@./bin/queue

build:
	@go build -o ./bin/queue
