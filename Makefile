.PHONY: build generate-proto generate-code

build: generate-proto generate-code
	@echo "Running Go build..."
	go build ./...
	go fmt ./...
	go vet ./...

generate-proto:
	@echo "Generating proto files..."
	protoc --proto_path=. \
		--go_out=./grackle/v1beta \
		--go_opt=module=github.com/evrblk/evrblk-go/grackle/v1beta \
		--go-grpc_out=./grackle/v1beta \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/grackle/v1beta \
		--go-vtproto_out=./grackle/v1beta \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/grackle/v1beta \
		./proto/grackle/v1beta/*.proto
	protoc --proto_path=. \
		--go_out=./banyan/v0 \
		--go_opt=module=github.com/evrblk/evrblk-go/banyan/v0 \
		--go-grpc_out=./banyan/v0 \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/banyan/v0 \
		--go-vtproto_out=./banyan/v0 \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/banyan/v0 \
		./proto/banyan/v0/*.proto
	protoc --proto_path=. \
		--go_out=./iam/v0 \
		--go_opt=module=github.com/evrblk/evrblk-go/iam/v0 \
		--go-grpc_out=./iam/v0 \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/iam/v0 \
		--go-vtproto_out=./iam/v0 \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/iam/v0 \
		./proto/iam/v0/*.proto
	protoc --proto_path=. \
		--go_out=./moab/v0 \
		--go_opt=module=github.com/evrblk/evrblk-go/moab/v0 \
		--go-grpc_out=./moab/v0 \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/moab/v0 \
		--go-vtproto_out=./moab/v0 \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/moab/v0 \
		./proto/moab/v0/*.proto
	protoc --proto_path=. \
		--go_out=./myaccount/v0 \
		--go_opt=module=github.com/evrblk/evrblk-go/myaccount/v0 \
		--go-grpc_out=./myaccount/v0 \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/myaccount/v0 \
		--go-vtproto_out=./myaccount/v0 \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/myaccount/v0 \
		./proto/myaccount/v0/*.proto

generate-code:
	@echo "Running code generate..."
	go run ./cmd/codegen \
		--service-name=Grackle \
		--go-package-path=github.com/evrblk/evrblk-go/grackle/v1beta \
		--go-package-name=grackle \
		--output-path=./grackle/v1beta/client.go \
		--proto-file-path=./proto/grackle/v1beta/api.proto
	go run ./cmd/codegen \
		--service-name=Banyan \
		--go-package-path=github.com/evrblk/evrblk-go/banyan/v0 \
		--go-package-name=banyan \
		--output-path=./banyan/v0/client.go \
		--proto-file-path=./proto/banyan/v0/api.proto
	go run ./cmd/codegen \
		--service-name=IAM \
		--go-package-path=github.com/evrblk/evrblk-go/iam/v0 \
		--go-package-name=iam \
		--output-path=./iam/v0/client.go \
		--proto-file-path=./proto/iam/v0/api.proto
	go run ./cmd/codegen \
		--service-name=Moab \
		--go-package-path=github.com/evrblk/evrblk-go/moab/v0 \
		--go-package-name=moab \
		--output-path=./moab/v0/client.go \
		--proto-file-path=./proto/moab/v0/api.proto
	go run ./cmd/codegen \
		--service-name=MyAccount \
		--go-package-path=github.com/evrblk/evrblk-go/myaccount/v0 \
		--go-package-name=myaccount \
		--output-path=./myaccount/v0/client.go \
		--proto-file-path=./proto/myaccount/v0/api.proto
