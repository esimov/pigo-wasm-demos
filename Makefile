ifeq ($(OS),Windows_NT)
    BROWSER = start
else
	UNAME := $(shell uname -s)
	ifeq ($(UNAME), Linux)
		BROWSER = xdg-open
	endif
	ifeq ($(UNAME), Darwin)
		BROWSER = open
	endif
endif

.PHONY: all clean serve

all: wasm serve

demo1: masquerade serve
demo2: pixelate serve
demo3: faceblur serve
demo4: triangulate serve

masquerade:
	cp -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ./js/
	GOOS=js GOARCH=wasm go build -o lib.wasm masquerade.go

pixelate:
	cp -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ./js/
	GOOS=js GOARCH=wasm go build -o lib.wasm pixelate.go

faceblur:
	cp -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ./js/
	GOOS=js GOARCH=wasm go build -o lib.wasm faceblur.go

triangulate:
	cp -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ./js/
	GOOS=js GOARCH=wasm go build -o lib.wasm triangulate.go

serve:
	$(BROWSER) 'http://localhost:5000'
	serve

clean:
	rm -f *.wasm

debug:
	@echo $(UNAME)
