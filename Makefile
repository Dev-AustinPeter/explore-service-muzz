.PHONY: gen test

gen:
	@protoc \
		--proto_path=. "explore-service.proto" \
		--go_out=genproto/explore --go_opt=paths=source_relative \
		--go-grpc_out=genproto/explore \
		--go-grpc_opt=paths=source_relative

test:
	go test -tags unit -v ./...