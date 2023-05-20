package chat

import (
	"context"
	"log"

	chatpb "github.com/DiarCode/grpc-chat-app/src/chat/gen"
	"github.com/DiarCode/grpc-chat-app/src/chat/services"
	"github.com/rabbitmq/amqp091-go"
)

type ChatServer struct {
	MessageQueue *amqp091.Channel
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

	// Continuously listen for new messages
	for {
		select {
		case d, ok := <-msgs:
			if !ok {
				// Channel closed
				return nil
			}
			message := &chatpb.Message{
				Username: d.Headers["username"].(string),
				Text:     string(d.Body),
			}
			if err := stream.Send(message); err != nil {
				return err
			}
		case <-stream.Context().Done():
			// Client disconnected
			log.Printf("User %s left the chat", req.Username)
			return nil
		}
	}
}

func (s *ChatServer) SendMessage(ctx context.Context, req *chatpb.Message) (*chatpb.Message, error) {
	log.Printf("Received message from %s: %s", req.Username, req.Text)

	err := services.PublishToQueue(
		s.MessageQueue,
		MESSAGE_QUEUE_NAME,
		[]byte(req.Text),
		amqp091.Table{"username": req.Username},
	)

	if err != nil {
		return nil, err
	}

	return &chatpb.Message{
		Username: req.Username,
		Text:     req.Text,
	}, nil
}
