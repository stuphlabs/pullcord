ROOTPKG = github.com/stuphlabs/pullcord
CONTAINERNAME = pullcord

binfiles = bin/pullcord bin/genhash

PKG = ./...
COVERMODE = set

cleanfiles = cover.html cover.out bin ${binfiles}
recursivecleanfiles = pkg.cov

.PHONY: all
all: test ${binfiles}

bin/%: cmd/%/*.go
	$(MAKE) get
	mkdir -p bin
	go build -v -o $@ ./cmd/$*

.PHONY: clean
clean:
	-rm -rv ${cleanfiles} `for file in ${recursivecleanfiles}; do \
		find . -name $${file}; \
	done`

container: test Dockerfile ${binfiles}
	docker build -t ${CONTAINERNAME} .

cover.html: cover.out
	go tool cover -html $< -o $@

cover.out: *.go */*.go
	go test -v -coverprofile $@ -covermode ${COVERMODE} ${PKG}

.PHONY: get
get:
	go get -t -u -v ${PKG}

.PHONY: test
test: cover.out cover.html

