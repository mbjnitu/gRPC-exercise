package chatserver

import (
	"log"
	"math/rand"
	"sync"
	"time"
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

var clientCount int = 0
var clientsThatReceivedMessage, clientsThatLeftChat []int

// define ChatService
func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {

	clientUniqueCode := rand.Intn(1e6)
	errch := make(chan error)

	clientCount++
	// receive messages - init a go routine
	go receiveFromStream(csi, clientUniqueCode, errch)

	// send messages - init a go routine
	go sendToStream(csi, clientUniqueCode, errch)

	return <-errch

}

// receive messages
func receiveFromStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	//implement a loop
	for {
		if contains(clientsThatLeftChat, clientUniqueCode_) {
			break
		}
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			errch_ <- err
		} else {

			messageHandleObject.mu.Lock()

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				ClientName:        mssg.Name,
				MessageBody:       mssg.Body,
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode_,
			})

			log.Printf("%v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

			messageHandleObject.mu.Unlock()

		}
	}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// send message
func sendToStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	//implement a loop
	for {

		//loop through messages in MQue
		for {

			time.Sleep(500 * time.Millisecond)

			if contains(clientsThatLeftChat, clientUniqueCode_) {
				break
			}

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			senderUniqueCode := messageHandleObject.MQue[0].ClientUniqueCode
			senderName4Client := messageHandleObject.MQue[0].ClientName
			message4Client := messageHandleObject.MQue[0].MessageBody

			messageHandleObject.mu.Unlock()

			//send message to designated client (do not send to the same client)
			if !contains(clientsThatReceivedMessage, senderUniqueCode) {
				clientsThatReceivedMessage = append(clientsThatReceivedMessage, senderUniqueCode)
				if len(clientsThatReceivedMessage) == clientCount {
					messageHandleObject.mu.Lock()
					clientsThatReceivedMessage = nil
					messageHandleObject.MQue = []messageUnit{}
					messageHandleObject.mu.Unlock()
				}
			}

			if !contains(clientsThatReceivedMessage, clientUniqueCode_) && clientCount > 1 {
				clientsThatReceivedMessage = append(clientsThatReceivedMessage, clientUniqueCode_)

				err := csi_.Send(&FromServer{Name: senderName4Client, Body: message4Client})

				if err != nil {
					errch_ <- err
				}

				messageHandleObject.mu.Lock()

				if len(clientsThatReceivedMessage) == clientCount && message4Client != "Left the chatroom" {
					clientsThatReceivedMessage = nil
					messageHandleObject.MQue = []messageUnit{}
				}

				messageHandleObject.mu.Unlock()

			}

			if len(clientsThatReceivedMessage) == clientCount && clientCount > 1 && message4Client == "Left the chatroom" && clientUniqueCode_ == senderUniqueCode {
				csi_.Send(&FromServer{Name: senderName4Client, Body: "leaveToken:1230123"})
				messageHandleObject.mu.Lock()
				clientsThatReceivedMessage = nil
				messageHandleObject.MQue = []messageUnit{}
				messageHandleObject.mu.Unlock()
				clientCount--
				clientsThatLeftChat = append(clientsThatLeftChat, senderUniqueCode)
			}

		}

		time.Sleep(100 * time.Millisecond)
	}
}
