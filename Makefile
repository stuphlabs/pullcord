ROOTPKG = github.com/stuphlabs/pullcord
CONTAINERNAME = pullcord

binfiles = bin/pullcord bin/genhash

PKG = ./...
COVERMODE = set
GO = go
DOCKER = docker

cleanfiles = bin cover.html cover.out ${binfiles}
recursive_cleanfiles =
dist_cleanfiles = .build_gopath
recursive_dist_cleanfiles =
maintainer_cleanfiles =
recursive_maintainer_cleanfiles =
go_build_files = .build_gopath/src
go_run = cd .build_gopath/src/${ROOTPKG} && GOPATH=$${PWD}/.build_gopath ${GO}

.PHONY: all
all: test ${binfiles}

bin/%: cmd/%/*.go ${go_build_files}
	mkdir -p bin
	${go_run} build -v -o $@ ./cmd/$*

.build_gopath/src:
	mkdir -p $@/`dirname ${ROOTPKG}`
	ln -s $${PWD} $@/${ROOTPKG}
	${go_run} get -t -v ${PKG}

.PHONY: clean
clean:
	-rm -rf ${cleanfiles}
	-rm -f `for file in ${recursive_cleanfiles}; do \
		find . -name $${file}; \
	done`

container: test Dockerfile ${binfiles}
	${DOCKER} build -t ${CONTAINERNAME} .

cover.html: cover.out ${go_build_files}
	${go_run} tool cover -html $< -o $@

cover.out: *.go */*.go ${go_build_files}
	${go_run} test -v -coverprofile $@ -covermode ${COVERMODE} ${PKG}

.PHONY: distclean
distclean: clean
	-rm -rf ${dist_cleanfiles}
	-rm -f `for file in ${recursive_dist_cleanfiles}; do \
		find . -name $${file}; \
	done`

.PHONY: maintainer_clean
maintainer_clean: clean distclean
	-rm -rf ${maintainer_cleanfiles}
	-rm -f `for file in ${recursive_maintainer_cleanfiles}; do \
		find . -name $${file}; \
	done`

.PHONY: test
test: cover.out cover.html

