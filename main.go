package main

import (
	"fmt"
	"github.com/anchamber/genetics-tank/db"
	pb "github.com/anchamber/genetics-tank/proto"
	"github.com/anchamber/genetics-tank/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	configuration := LoadConfiguration()

	addr := fmt.Sprintf(":%s", configuration.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	s := grpc.NewServer()
	pb.RegisterTankServiceServer(s, service.New(db.NewMockDB(nil)))

	// Serve gRPC Server
	log.Printf("Starting gRPC server %s\n", addr)
	log.Fatal(s.Serve(lis))
}
