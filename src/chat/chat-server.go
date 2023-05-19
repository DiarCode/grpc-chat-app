package chat

import (
	"log"
	"math/rand"
	"sync"
	"time"

	chatpb "github.com/DiarCode/grpc-chat-app/src/chat/gen"
)

type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageHandleObject = messageHandle{}

type ChatServer struct {
}

// define ChatService
func (is *ChatServer) ChatService(csi chatpb.ChatService_BroadcastMessageServer) error {

	clientUniqueCode := rand.Intn(1e6)
	err := make(chan error)

	// receive messages - init a go routine
	go receiveFromStream(csi, clientUniqueCode, err)

	// send messages - init a go routine
	go sendToStream(csi, clientUniqueCode, err)

	return <-err

}

// receive messages
func receiveFromStream(csi_ chatpb.ChatService_BroadcastMessageServer, clientUniqueCode_ int, err_ chan error) {

	//implement a loop
	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			err_ <- err
		} else {

			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:        mssg.Username,
				MessageBody:       mssg.Message,
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode_,
			})

			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

			messageHandleObject.mu.Unlock()
		}
	}
}

// send message
func sendToStream(csi_ chatpb.ChatService_BroadcastMessageServer, clientUniqueCode_ int, err_ chan error) {

	//implement a loop
	for {

		//loop through messages in MQue
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			senderUniqueCode := messageHandleObject.MQue[0].ClientUniqueCode
			senderName := messageHandleObject.MQue[0].ClientName
			message := messageHandleObject.MQue[0].MessageBody

			messageHandleObject.mu.Unlock()

			//send message to designated client (do not send to the same client)
			if senderUniqueCode != clientUniqueCode_ {

				err := csi_.Send(&chatpb.Message{Username: senderName, Message: message})

				if err != nil {
					err_ <- err
				}

				messageHandleObject.mu.Lock()

				if len(messageHandleObject.MQue) > 1 {
					messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete the message at index 0 after sending to receiver
				} else {
					messageHandleObject.MQue = []messageUnit{}
				}

				messageHandleObject.mu.Unlock()
			}

		}

		time.Sleep(100 * time.Millisecond)
	}
}
