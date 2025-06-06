.PHONY: test
test: TEST_RUN?=^.*$$
test: TEST_VERBOSE?=
test:
	go test \
		$(if $(TEST_VERBOSE),-v,) \
		-race \
        -timeout 1h \
		-coverprofile cp.out \
		-run '$(TEST_RUN)' \
		./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint run
