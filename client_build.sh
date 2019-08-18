OVERWRITE_SERVICE=0
REBOOT=0
CLIENT=$1
CBUILD=1
sudo apt install gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf
mkdir bin/ > /dev/null 2>&1;
echo 'building...';
export GOOS=linux
export GOARCH=arm
export GOARM=5
if [[ $CBUILD -eq 1 ]]; then
	export CC=arm-linux-gnueabihf-gcc
	export CXX=arm-linux-gnueabihf-g++
	export PKG_CONFIG="no"
	export CGO_ENABLED=1
	LIBROOT="/home/user/go/src/git.circuitco.de/self/greyhouse/rpi_libs/"
	export CGO_CPPFLAGS="-I$LIBROOT/include"
	LDEPS="-lz -ljpeg -lpng16 -ltiff -ldc1394 -lavcodec -lavformat"
	export CGO_LDFLAGS="-L$LIBROOT/lib -L$LIBROOT/sharelib -lopencv_core -lopencv_imgproc -lopencv_imgcodecs -lopencv_videoio $LDEPS"
	go build -tags customenv client.go 2>buildlog
else
	go build client.go 2>buildlog
fi
er=$?
if [ $er -ne 0 ]; then
	echo "failed to build"

	# Look for the desired lines
	python procdeps.py
	exit 1
fi
mv client bin/greyclient
echo 'built'
if [ ! -z $1 ]; then
	echo 'copying to clients'
	for var in "$@"; do
		echo "Copying to $var"
		rsync bin/greyclient $var:~/
	done
	for var in "$@"; do
		ssh cherry@$var cat /etc/systemd/system/greyclient.service > /dev/null
		ret=$?
		cmd=""
		if [[ $ret -ne 0 || $OVERWRITE_SERVICE -eq 1 ]]; then
			# service does not exist, configure
			echo "Configuring service for $var (ret $ret over $OVERWRITE_SERVICE)"
			rsync greyclient.service $var:/tmp/greyclient.service
			cmd='sudo mv /tmp/greyclient.service /etc/systemd/system/&&sudo systemctl daemon-reload&&sudo systemctl start greyclient.service'
			if [[ $? -ne 0 ]]; then
				echo "warning- failed to start for $var"
			fi
		else
			echo "Restarting service"
			# service exists, restart
			cmd='sudo systemctl restart greyclient.service'
		fi
		if [[ $cmd != "" ]]; then
			if [[ $REBOOT -ne 0 ]]; then
				echo "Cmd will reboot upon completion."
				cmd="$cmd&&sudo reboot&&exit"
			fi
			ssh cherry@$var -t $cmd
		fi
	done
	echo 'copy complete'
else
	echo "say './build.sh hostname' for rsync"
fi
