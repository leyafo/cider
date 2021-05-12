.PHONY: run
run:
	@go run main.go


.PHONY: build
build:
	@go build -o cider main.go


clean:
	@rm -f cider .meta
	@rm cider_*

platforms := $(windows linux darwin)
release:
	@for v in windows linux darwin ; do \
		GOOS=$$v GOARCH=amd64 go build -o cider_$${v}_amd64 *.go ; \
		zip -ur cider_$${v}_amd64.zip cider_$${v}_amd64 templates ; \
		GOOS=$$v GOARCH=386 go build -o cider_$${v}_386 *.go ; \
		zip -ur cider_$${v}_386.zip cider_$${v}_386 templates ; \
    done
