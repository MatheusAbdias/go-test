test:
	@go test -v -tags=integration ./...

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

server:
	@go run ./cmd/web

clean:
	@$(RM) *.out

.PHONY: test test-coverage clean server