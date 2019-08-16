PKG := .

# Build
all: codecheck test

test:
	go test -v -race

# Needs reflex package to be installed github.com/cespare/reflex
# Must be run with sudo
doc:
	@mkdir -p /tmp/tmpgoroot/doc
	@rm -rf /tmp/tmpgopath/src/github.com/colinc86/parallel
	@mkdir -p /tmp/tmpgopath/src/github.com/colinc86/parallel
	@tar -c --exclude='.git' --exclude='tmp' . | tar -x -C /tmp/tmpgopath/src/github.com/colinc86/parallel
	@echo "Starting GODOC Server"
	@reflex -g '*.go' make tarDir & GOROOT=/tmp/tmpgoroot/ GOPATH=/tmp/tmpgopath/ godoc -http=localhost:6060
	
codecheck: fmt lint vet

fmt:
	@echo "+ go fmt"
	go fmt $(PKG)

# Needs lint package to be installed golang.org/x/lint/golint
lint:
	@echo "+ go lint"
	golint -min_confidence=0.3 $(PKG)/...

vet:
	@echo "+ go vet"
	go vet $(PKG)/...
