package main

import (
	"fmt"
	"io"
	"net"
	proto "talktunnel"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	proto.UnimplementedChatServiceServer
	clients map[uint32]map[string]proto.ChatService_ChatStreamServer
}

func (s *server) registerClient(stream proto.ChatService_ChatStreamServer, msg *proto.Message) {
	if _, ok := s.clients[msg.GetRoomId()]; !ok {
		s.clients[msg.GetRoomId()] = make(map[string]proto.ChatService_ChatStreamServer)
	}
	messages := getMessages(msg.GetRoomId())
	if len(messages) > 0 {
		for _, msg := range messages {
			stream.Send(msg)
		}
	}

	stream.Send(&proto.Message{
		Text: fmt.Sprintf("hola %s, bienvenido a la sala %d!", msg.GetName(), msg.GetRoomId()),
		Name: "server",
	})

	s.clients[msg.GetRoomId()][msg.GetName()] = stream
}

func (s *server) handleMessage(stream proto.ChatService_ChatStreamServer, msg *proto.Message) {
	addMessage(msg)
	for _, s := range s.clients[msg.GetRoomId()] {
		s.Send(msg)
	}
}

func (s *server) ChatStream(stream proto.ChatService_ChatStreamServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			if status.Code(err) != codes.Canceled {
				fmt.Println("Failed to receive message:", err)
			}
			continue
		}
		if msg == nil {
			fmt.Println("Received nil message")
			continue
		}
		if msg.GetText() == "" {
			fmt.Println("Received '' message")
			continue
		}

		if msg.GetText() == "register" {
			fmt.Println("Register: ", msg)
			s.registerClient(stream, msg)
		} else {
			fmt.Println("New message: ", msg)
			s.handleMessage(stream, msg)
		}
	}
}

func main() {
	startDB()
	lis, _ := net.Listen("tcp", ":50051")
	s := grpc.NewServer()
	serverRef := &server{clients: make(map[uint32]map[string]proto.ChatService_ChatStreamServer)}
	proto.RegisterChatServiceServer(s, serverRef)
	s.Serve(lis)
}
