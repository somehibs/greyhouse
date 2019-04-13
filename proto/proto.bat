cd proto
protoc --go_out=plugins=grpc:..\api person.proto node.proto presence.proto rules.proto
cd ..