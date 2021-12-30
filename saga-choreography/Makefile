docker:
	docker build -t wallet --target wallet .
	docker build -t bank --target bank .

.PHONY: test
test:
	go test -v ./test