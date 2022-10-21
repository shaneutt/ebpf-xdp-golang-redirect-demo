LIBBPF ?= libbpf/src
CLANG ?= clang
CFLAGS := -O2 -g -Wall -Werror -Wno-unused-value -Wno-pointer-sign -Wcompare-distinct-pointer-types -I$(LIBBPF) $(CFLAGS)

all: build

.PHONY:
clean:
	rm -f bpf_bpfeb.o
	rm -f bpf_bpfel.o
	rm -f demo

.PHONY: generate
generate: export BPF_CLANG := $(CLANG)
generate: export BPF_CFLAGS := $(CFLAGS)
generate:
	go generate ./...

.PHONY: build
build: generate
	go build -o demo
