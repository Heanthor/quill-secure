run-leader:
	go run leader/main.go leader/net.go

deviceID := 0
run-node:
	go run node/main.go node/poll.go -deviceID=${deviceID}

sync:
	bash sync.sh