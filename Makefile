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

%.wasm: %.go
	cp -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ./js/
	GOOS=js GOARCH=wasm go generate
	GOOS=js GOARCH=wasm go build -o lib.wasm "$<"

demo1: masquerade.wasm serve
demo2: faceblur.wasm serve
demo3: pixelate.wasm serve
demo4: triangulate.wasm serve
demo5: facemask.wasm serve

serve:
	$(BROWSER) 'http://localhost:5000'
	go run server/init.go

clean:
	rm -f *.wasm

debug:
	@echo $(UNAME)
