TEST_FILES?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: verify
verify: fmt gobuild
all: k8s fmt gobuild

k8s:
	operator-sdk generate k8s

fmt:
	go fmt ./...
gobuild:
	go build -o asm-se-syncer ./cmd/...

vet:
	@echo "go vet ./..."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi
