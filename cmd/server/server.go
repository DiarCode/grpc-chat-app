package main

import (
	"log"
	"net"

	"github.com/DiarCode/grpc-chat-app/src/chat"
	chatpb "github.com/DiarCode/grpc-chat-app/src/chat/gen"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer channel.Close()
	log.Println("RabbitMQ channel created")

	server := grpc.NewServer()
	chatpb.RegisterChatServiceServer(server, &chat.ChatServer{MessageQueue: channel})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Server is running on port %v", port)

	err = server.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
