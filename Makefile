

.PHONY: run
run:
	@go run *.go


.PHONY: build
build:
	@go build -o cider *.go


deploy: build
	@cp -f cider /mnt/c/Users/Tank/code/blog_data/

clean:
	@rm -f cider .meta
