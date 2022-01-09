generate-mocks:
ifeq ($(shell which mockgen),)
	@echo "mockgen not found, install dependencies"
	go github.com/golang/mock/mockgen@latest
endif
	go generate ./...
