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
		--go_out=./banyan/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/banyan/preview \
		--go-grpc_out=./banyan/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/banyan/preview \
		--go-vtproto_out=./banyan/preview \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/banyan/preview \
		./proto/banyan/preview/*.proto
	protoc --proto_path=. \
		--go_out=./iam/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/iam/preview \
		--go-grpc_out=./iam/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/iam/preview \
		--go-vtproto_out=./iam/preview \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/iam/preview \
		./proto/iam/preview/*.proto
	protoc --proto_path=. \
		--go_out=./moab/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/moab/preview \
		--go-grpc_out=./moab/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/moab/preview \
		--go-vtproto_out=./moab/preview \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/moab/preview \
		./proto/moab/preview/*.proto
	protoc --proto_path=. \
		--go_out=./myaccount/preview \
		--go_opt=module=github.com/evrblk/evrblk-go/myaccount/preview \
		--go-grpc_out=./myaccount/preview \
		--go-grpc_opt=module=github.com/evrblk/evrblk-go/myaccount/preview \
		--go-vtproto_out=./myaccount/preview \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-vtproto_opt=module=github.com/evrblk/evrblk-go/myaccount/preview \
		./proto/myaccount/preview/*.proto

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
		--go-package-path=github.com/evrblk/evrblk-go/banyan/preview \
		--go-package-name=banyan \
		--output-path=./banyan/preview/client.go \
		--proto-file-path=./proto/banyan/preview/api.proto
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
