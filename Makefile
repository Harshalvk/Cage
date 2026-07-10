lint:
	golangci-lint run
	
fmt:
	gofmt -w .
	goimports -w .