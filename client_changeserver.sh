if [[ -z $1 || -z $2 ]]; then
	echo "cannot continue - first arg is ip, second+ args are clients"
else
	for client in "$@"; do
		if [[ $client == $1 ]]; then
			continue
		fi
		ssh cherry@$client sed -i "s/[0-9\.]*:9999/$1:9999/" ./client.json
		ssh cherry@$client cat ./client.json
		ssh cherry@$client -t sudo systemctl restart greyclient.service
	done
fi
