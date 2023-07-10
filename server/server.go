package main

import (
	"fmt"
	"io"
	"net"
	proto "talktunnel"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedChatServiceServer
	clients map[uint32]map[string]proto.ChatService_ChatStreamServer
}

func (s *server) ChatStream(stream proto.ChatService_ChatStreamServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if msg.GetText() == "register" {
			stream.Send(&proto.Message{
				Text: fmt.Sprintf("%s, registrado correctamente!", msg.GetName()),
				Name: "server",
			})

			if _, ok := s.clients[msg.GetRoomId()]; !ok {
				s.clients[msg.GetRoomId()] = make(map[string]proto.ChatService_ChatStreamServer)
			}
			messages := getMessages(msg.GetRoomId())
			if len(messages) > 0 {
				for _, msg := range messages {
					stream.Send(msg)
				}
			}

			s.clients[msg.GetRoomId()][msg.GetName()] = stream
		} else if (msg != nil) && (msg.GetText() != "") {
			addMessage(msg)
			for _, s := range s.clients[msg.GetRoomId()] {
				s.Send(msg)
			}
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
