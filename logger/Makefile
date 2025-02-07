.PHONY: gen
gen:
	@protoc \
		--proto_path=proto "proto/logs.proto" \
		--go_out=proto/genproto --go_opt=paths=source_relative \
		--go-grpc_out=proto/genproto --go-grpc_opt=paths=source_relative
