.PHONY: build clean serve

TINYGO = tinygo
TINYGOROOT ?= /tmp/tinygo
WASM_SRC = wasm/wasm.go
WASM_OUT = web/float_conv.wasm
HTML = web/index.html
JS = web/wasm_exec.js

build: $(WASM_OUT) $(JS)
	@echo "Build complete: $(WASM_OUT)"

$(WASM_OUT): $(WASM_SRC)
	cd wasm && $(TINYGO) build -target=wasm -o ../$(WASM_OUT) .
	@cp $(TINYGOROOT)/targets/wasm_exec.js web/

$(JS):
	@cp $(TINYGOROOT)/targets/wasm_exec.js web/

clean:
	rm -f $(WASM_OUT)

serve:
	cd web && python3 -m http.server 8080

.DEFAULT_GOAL := build