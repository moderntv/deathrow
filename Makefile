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


