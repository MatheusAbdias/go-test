test:
	@go test -v ./...

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

clean:
	@$(RM) *.out

.PHONY: test test-coverage clean