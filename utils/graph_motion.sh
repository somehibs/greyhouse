go build -o grapher motion.go; while true; do scp lounge:~/motion.csv . && ./grapher; sleep 0.25; done
