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
		rsync bin/greyclient $var:~/
	done
	echo 'copy complete'
else
	echo "say './build.sh hostname' for rsync"
fi
