wasm:
	@GOARCH=wasm GOOS=js go build -o "./static/client.wasm" ../client/main.go
	@echo "wasm done"
run: wasm
	@go build -o server main.go
	@./server