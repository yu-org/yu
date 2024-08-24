
test:
	go test -v ./tests/single_node_test.go

benchTPS:
	go test -v ./tests/bench_transfer_test.go

check-mod-tidy:
	@go mod tidy
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Changes detected after running go mod tidy. Please run 'go mod tidy' locally and commit the changes."; \
		git status; \
		exit 1; \
	else \
		echo "No changes detected after running go mod tidy."; \
	fi