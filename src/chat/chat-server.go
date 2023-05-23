package chat

import (
	"context"
	"log"

	chatpb "github.com/DiarCode/grpc-chat-app/src/chat/gen"
	"github.com/DiarCode/grpc-chat-app/src/chat/services"
	"github.com/streadway/amqp"
)

type ChatServer struct {
	MessageQueue *amqp.Channel
}

const (
	MESSAGE_QUEUE_NAME = "messageQueue"
)

func (s *ChatServer) JoinStream(req *chatpb.JoinRequest, stream chatpb.ChatService_JoinStreamServer) error {
	log.Printf("User %s joined the chat", req.Username)

	queue, err := services.DeclareQueue(s.MessageQueue, MESSAGE_QUEUE_NAME)
	if err != nil {
		return err
	}

	msgs, err := services.ConsumeFromQueue(s.MessageQueue, queue.Name)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-msgs:
				message := &chatpb.Message{
					Username: msg.Headers["username"].(string),
					Text:     string(msg.Body),
				}

				if err := stream.Send(message); err != nil {
					log.Printf("Failed to send message to client: %v", err)
					return
				}
			case <-stream.Context().Done():
				log.Printf("User %s left the chat", req.Username)
				return
			}
		}
	}()

	// Wait for the client to close the stream
	<-stream.Context().Done()

	return nil

}

func (s *ChatServer) SendMessage(ctx context.Context, req *chatpb.Message) (*chatpb.EmptyResponse, error) {
	log.Printf("Received message from %s: %s", req.Username, req.Text)

	err := services.PublishToQueue(
		s.MessageQueue,
		MESSAGE_QUEUE_NAME,
		[]byte(req.Text),
		amqp.Table{"username": req.Username},
	)

	if err != nil {
		return nil, err
	}

	return &chatpb.EmptyResponse{}, nil
}
