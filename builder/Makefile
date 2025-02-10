test: 
	$(MAKE) logs
	go test -coverprofile=coverage.out -covermode=set $(go list ./...) -v ./...
	$(MAKE) fix 

coverage:
	go tool cover -html=coverage.out

clean:
	find . -name "*.log" -delete
	find . -name "*.out" -delete
	find . -name "*.db" -delete
	find . -name "logs" -type d -empty -delete
	find . -name "coverage.out" -delete
	find . -name "covprofile" -delete

.PHONY: logs
logs:
	cd ../scripts && python3 find_logs.py

.PHONY: fix
fix:
	cd ../scripts && python3 find_todos.py


.PHONY: migration 
migration:
	migrate create -ext sql -dir migrations -seq $(name)
