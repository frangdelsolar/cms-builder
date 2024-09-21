test:
	$(MAKE) logs
	go test -coverprofile=coverage.out -covermode=set $(go list ./...) -v ./...

coverage:
	go tool cover -html=coverage.out

clean:
	find . -name "*.log" -delete
	find . -name "*.out" -delete
	find . -name "*.db" -delete
	find . -name "logs" -type d -empty -delete

.PHONY: logs
logs:
	cd ../scripts && python3 find_logs.py