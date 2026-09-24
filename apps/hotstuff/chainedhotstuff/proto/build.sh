protoc -I . --go_opt=paths=source_relative --go_out=plugins=grpc:./ chainedbft.proto
