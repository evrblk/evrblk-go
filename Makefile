.PHONY: build generate-proto generate-code

build: generate-proto generate-code
	@echo "Running Go build..."
	go build ./...
	go fmt ./...
	go vet ./...

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

generate-code:
	@echo "Running code generate..."
	go run ./cmd/codegen \
		--service-name=Grackle \
		--go-package-path=github.com/evrblk/evrblk-go/grackle/preview \
		--go-package-name=grackle \
		--output-path=./grackle/preview/client.go \
		--proto-file-path=./proto/grackle/preview/api.proto
	go run ./cmd/codegen \
		--service-name=IAM \
		--go-package-path=github.com/evrblk/evrblk-go/iam/preview \
		--go-package-name=iam \
		--output-path=./iam/preview/client.go \
		--proto-file-path=./proto/iam/preview/api.proto
	go run ./cmd/codegen \
		--service-name=Moab \
		--go-package-path=github.com/evrblk/evrblk-go/moab/preview \
		--go-package-name=moab \
		--output-path=./moab/preview/client.go \
		--proto-file-path=./proto/moab/preview/api.proto
	go run ./cmd/codegen \
		--service-name=MyAccount \
		--go-package-path=github.com/evrblk/evrblk-go/myaccount/preview \
		--go-package-name=myaccount \
		--output-path=./myaccount/preview/client.go \
		--proto-file-path=./proto/myaccount/preview/api.proto
