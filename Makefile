protos:
	protoc --go_out=. --go-grpc_out=. protos/**/*.proto


client:
	go run ./cmd/client $(username)


setup:
	docker run -d --hostname my-rabbit --name some-rabbit -p 15672:15672 -p 5672:5672 rabbitmq:3
	go run ./cmd/server




