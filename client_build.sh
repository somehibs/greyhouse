OVERWRITE_SERVICE=0
REBOOT=0
CLIENT=$1
mkdir bin/ > /dev/null 2>&1;
echo 'building...';
export GOOS=linux
export GOARCH=arm
export GOARM=5
go build client.go
er=$?
if [ $er -ne 0 ]; then
	echo "failed to build"
	exit $er
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
