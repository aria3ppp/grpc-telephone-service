package main

import (
	"log"
	"net"

	"github.com/aria3ppp/grpc-telephone-service/gapi"
	"github.com/aria3ppp/grpc-telephone-service/pb"
	"google.golang.org/grpc"
)

func main() {
	l, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	server := grpc.NewServer()
	pb.RegisterTelephoneServer(server, gapi.NewServer())

	log.Printf("grpc server is running on %s...", l.Addr().String())
	if err := server.Serve(l); err != nil {
		log.Fatal("failed to serve:", err)
	}
}
