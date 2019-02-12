setup:
	@echo "Donwloading dep to use as package manager"
	@go get -u github.com/golang/dep/cmd/dep
	@echo "Downloading dependencies"
	@dep ensure
	@echo "Downloading gometalinter for linting"
	@go get -u github.com/alecthomas/gometalinter
	@echo "Downloading gocov for code coverage analyzes"
	@go get -u github.com/axw/gocov/gocov
	@echo "Downloading reflex for watching over the files"
	@go get -u github.com/cespare/reflex

build: setup
	@go build -o ./bin/pusher-tester main.go

start:
	@./bin/pusher-tester
