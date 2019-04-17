mkdir bin/ > /dev/null 2>&1; echo 'building...'; go build client.go; mv client bin/greyhouse; echo 'built'
