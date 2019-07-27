CLIENT=$1
mkdir bin/ > /dev/null 2>&1;
echo 'building...';
go build client.go
er=$?
if [ $er -ne 0 ]; then
	echo "failed to build"
	exit $er
fi
mv client bin/greyhouse
echo 'built'
if [ ! -z $1 ]; then
	echo 'copying to clients'
	rsync bin/greyhouse $CLIENT:~/
	echo 'copy complete'
else
	echo "say './build.sh hostname' for rsync"
fi
