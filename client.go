package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"grpcChatServer/chatserver"

	"google.golang.org/grpc"
)

var Lamport int = 0

func main() {
	//connect to grpc server
	conn, err := grpc.Dial("localhost:5000", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Faile to conncet to gRPC server :: %v", err)
	}
	defer conn.Close()

	//call ChatService to create a stream
	client := chatserver.NewServicesClient(conn)

	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}

	// implement communication with gRPC server
	ch := clienthandle{stream: stream}
	ch.clientConfig()
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

		Lamport = chatserver.IncrementLamport(Lamport) //Sending a message will increase the Lamport time

		clientMessageBox := &chatserver.FromClient{
			Name: ch.clientName,
			Body: clientMessage + " | " + strconv.Itoa(Lamport),
		}
		if len(clientMessage) < 128 {
			err = ch.stream.Send(clientMessageBox)
		} else {
			log.Printf("Your message is too long u dumb fuck")
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

		receivedLamport, BodyWOLamport := chatserver.SplitLamport(mssg.Body)
		Lamport = chatserver.SyncLamport(Lamport, receivedLamport)
		Lamport = chatserver.IncrementLamport(Lamport) //Receiving a message will increase the Lamport time

		if err != nil {
			log.Printf("Error in receiving message from server :: %v", err)
		}

		//print message to console
		fmt.Printf("%s : %s \n", mssg.Name, BodyWOLamport+" | "+strconv.Itoa(Lamport))
	}
}

func (ch *clienthandle) sendJoinMessage(name string) {

	Lamport = chatserver.IncrementLamport(Lamport) //Sending a message will increase the Lamport time

	clientMessageBox := &chatserver.FromClient{
		Name: name,
		Body: "Joined the chatroom | " + strconv.Itoa(Lamport),
	}
	ch.stream.Send(clientMessageBox)
}
