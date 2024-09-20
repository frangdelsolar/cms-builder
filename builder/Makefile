test:
	go test -coverprofile=coverage.out -covermode=set $(go list ./...) -v ./...

coverage:
	go tool cover -html=coverage.out