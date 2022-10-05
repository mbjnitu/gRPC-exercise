package main

import (
	gRPC "https://github.com/mbjnitu/gRPC-exercise/proto"
	"google.golang.org/grpc"
)

type Server struct {
	// an interface that the server needs to have
	gRPC.UnimplementedTemplateServer

	// here you can impliment other fields that you want
}
