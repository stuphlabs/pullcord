# Basic config to be filled in

ROOTPKG = github.com/stuphlabs/pullcord
CONTAINERNAME = pullcord

binfiles = bin/pullcord bin/genhash

# Most things below here won't need to change

PKG = ./...
COVERMODE = set

cleanfiles = all.cov cover.html bin ${binfiles}
recursivecleanfiles = pkg.cov

.PHONY: all
all: test ${binfiles}

.PHONY: all.cov
all.cov:
	echo "mode: ${COVERMODE}" > $@
	$(MAKE) `go list ${PKG} | sed 's#^${ROOTPKG}\(/\?.*\)$$#\.\1/pkg.cov#'`
	go list ${PKG} | sed 's#^${ROOTPKG}\(/\?.*\)$$#\.\1/pkg.cov#' \
	| while read pkgcov; do \
		tail -n +2 $${pkgcov} >> $@; \
	done

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

cover.html: all.cov
	go tool cover -html $< -o $@

.PHONY: get
get:
	go get -t -v ./...

pkg.cov: *.go
	$(MAKE) get
	go test -v -coverprofile $@ -covermode ${COVERMODE} .

%/pkg.cov: %/*.go
	$(MAKE) get
	go test -v -coverprofile $@ -covermode ${COVERMODE} ./$*

.PHONY: test
test: all.cov cover.html

