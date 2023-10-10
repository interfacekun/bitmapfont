run:
	go run -tags "chinese"  example/zh/main.go
facezh:
	go run -tags=generate ./internal/gen -chinese -output ./data/facezh.bin
