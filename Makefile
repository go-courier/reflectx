fmt:
	goimports -w -l .

test:
	go test -v -race ./...


bench:
	go test -test.bench=. -test.benchmem ./...

cover:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...