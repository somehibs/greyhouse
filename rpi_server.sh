CLIENT=$1
mkdir bin/ > /dev/null 2>&1;
echo 'building...';
export GOOS=linux
export GOARCH=arm
export GOARM=5
go build server.go
er=$?
if [ $er -ne 0 ]; then
	echo "failed to build"
	exit $er
fi
mv server bin/greyserver
echo 'built'
if [ ! -z $1 ]; then
	echo 'copying to clients'
	rsync bin/greyserver $CLIENT:~/
	ssh $CLIENT mkdir web
	rsync -r web/tpl/ $CLIENT:~/web/tpl/
	echo 'copy complete'
else
	echo "say './build.sh hostname' for rsync"
fi
