export GO111MODULE=off
if [[ ! -d /usr/local/go && ! -d /opt/go ]]; then
	echo "Go not installed. Assuming rpi, installing armv6"
	wget https://dl.google.com/go/go1.12.4.linux-armv6l.tar.gz
	sudo tar -C /usr/local -xzvf go1.12.4.linux-armv6l.tar.gz
else
	echo "Go OK"
fi

echo "Installing dependencies."
sudo apt install -y git wget protobuf-compiler
echo
if [[ ! -f /usr/local/bin/pigpiod ]]; then
	echo "Installing pi GPIO"
	wget abyz.me.uk/rpi/pigpio/pigpio.zip
	unzip pigpio.zip
	cd PIGPIO
	make
	sudo make install
else
	echo "pi GPIO detected"
fi
echo "Updating protoc-gen-go and greyhouse"
go get -u github.com/golang/protobuf/protoc-gen-go
go get git.circuitco.de/self/greyhouse
echo "updating tensorflow"
wget https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz
sudo tar -C /usr/local -xzf ./libtensorflow-cpu-linux-x86_64-1.14.0.tar.gz
rm libtensorflow-cpu-linux-*.tar.gz*
