setup:
	go get \
		github.com/laher/goxc \
		github.com/tcnksm/ghr \
		golang.org/x/lint/golint
	go get -d -t ./...

lint: setup
	go vet ./...
	golint -set_exit_status ./...

.PHONY: setup lint
