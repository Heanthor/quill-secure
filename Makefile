run-leader:
	go run $$(ls -1 leader/*.go | grep -v _test.go)

deviceID := 0
run-node:
	go run node/main.go node/poll.go -deviceID=${deviceID}

sync:
	bash sync.sh

build-node-prod:
	env GOOS=linux GOARCH=arm GOARM=5 go build -ldflags="-s -w" -o bin/node node/*.go

build-leader-prod:
	env GOOS=linux go build -ldflags="-s -w" -o bin/leader leader/main.go leader/net.go

deploy-leader:
	bash deploy/deploy_leader.sh

deploy-node: build-node-prod
	bash deploy/deploy_node.sh
