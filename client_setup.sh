sudo apt install -y git wget protobuf-compiler
wget https://dl.google.com/go/go1.12.4.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzvf go1.12.4.linux-armv6l.tar.gz
go get -u github.com/golang/protobuf/protoc-gen-go
go get git.circuitco.de/self/greyhouse
