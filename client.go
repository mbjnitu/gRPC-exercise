package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"grpcChatServer/chatserver"

	"google.golang.org/grpc"
)

func main() {
	//connect to grpc server
	conn, err := grpc.Dial("localhost:5000", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Faile to conncet to gRPC server :: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}

	//call ChatService to create a stream
	client := chatserver.NewServicesClient(conn)

	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}

	// implement communication with gRPC server
	ch := clienthandle{stream: stream}
	ch.clientName = strings.Trim(name, "\r\n")
	ch.sendJoinMessage(name)
	//ch.clientConfig()
	go ch.sendMessage()
	go ch.receiveMessage()

	//blocker
	bl := make(chan bool)
	<-bl

}

// clienthandle
type clienthandle struct {
	stream     chatserver.Services_ChatServiceClient
	clientName string
}

func (ch *clienthandle) clientConfig() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
	ch.sendJoinMessage(name)
}

// send message
func (ch *clienthandle) sendMessage() {

	// create a loop
	for {

		reader := bufio.NewReader(os.Stdin)
		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf(" Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")
		if clientMessage == "/leave" {
			clientMessage = "Left the chatroom"
			clientLeaveChatMessage := &chatserver.FromClient{
				Name: ch.clientName,
				Body: clientMessage,
			}
			err = ch.stream.Send(clientLeaveChatMessage)
			time.Sleep(1000 * time.Millisecond)
			os.Exit(0)
		} else {
			clientMessageBox := &chatserver.FromClient{
				Name: ch.clientName,
				Body: clientMessage,
			}
			if len(clientMessage) < 128 {
				err = ch.stream.Send(clientMessageBox)
			} else {
				log.Printf("Your message is too long u dumb dumb")
			}
		}

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}

	}

}

// receive message
func (ch *clienthandle) receiveMessage() {

	//create a loop
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in receiving message from server :: %v", err)
		}

		//print message to console
		fmt.Printf("%s : %s \n", mssg.Name, mssg.Body)

	}
}

func (ch *clienthandle) sendJoinMessage(name string) {
	clientMessageBox := &chatserver.FromClient{
		Name: name,
		Body: "Joined the chatroom - " + time.Now().Format("2006.01.02 15:04:05"),
	}
	ch.stream.Send(clientMessageBox)
}
