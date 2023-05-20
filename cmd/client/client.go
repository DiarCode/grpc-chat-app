package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	chatpb "github.com/DiarCode/grpc-chat-app/src/chat/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func receiveMessages(stream chatpb.ChatService_JoinStreamClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Chat stream closed")
				return
			}
			log.Fatalf("Failed to receive a message: %v", err)
		}
		log.Printf("[%s]: %s", msg.Username, msg.Text)
	}
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to the server: %v", err)
	}
	defer conn.Close()

	client := chatpb.NewChatServiceClient(conn)

	username := os.Args[1]
	log.Printf("======== Welcome, %s! ========", username)

	// Join the chat stream
	stream, err := client.JoinStream(context.Background(), &chatpb.JoinRequest{Username: username})
	if err != nil {
		log.Fatalf("Failed to join the chat stream: %v", err)
	}

	go receiveMessages(stream)

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Failed to read user input: %v", err)
			continue
		}

		message := &chatpb.Message{
			Username: username,
			Text:     input,
		}

		_, err = client.SendMessage(context.Background(), message)
		if err != nil {
			log.Fatalf("Failed to send a message: %v", err)
		}
	}
}
