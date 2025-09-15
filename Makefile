.PHONY: build generate-proto

build: generate-proto
	@echo "Running code generate..."
	go run ./cmd/codegen
	@echo "Running Go build..."
	go build ./...

generate-proto:
	@echo "Generating proto files..."
	protoc --proto_path=. \
		--go_out=./grackle/preview \
		--go-grpc_out=./grackle/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/grackle/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/grackle/preview \
		./proto/grackle/preview/*.proto
	protoc --proto_path=. \
		--go_out=./iam/preview \
		--go-grpc_out=./iam/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/iam/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/iam/preview \
		./proto/iam/preview/*.proto
	protoc --proto_path=. \
		--go_out=./moab/preview \
		--go-grpc_out=./moab/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/moab/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/moab/preview \
		./proto/moab/preview/*.proto
	protoc --proto_path=. \
		--go_out=./myaccount/preview \
		--go-grpc_out=./myaccount/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/myaccount/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/myaccount/preview \
		./proto/myaccount/preview/*.proto
