package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	proto "talktunnel"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	client := proto.NewChatServiceClient(conn)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("username:")
	name, _ := reader.ReadString('\n')
	name = strings.Trim(name, "\r\n")

	fmt.Print("roomId:")
	roomIdStr, _ := reader.ReadString('\n')
	roomIdStr = strings.Trim(roomIdStr, "\r\n")
	roomId, _ := strconv.ParseUint(roomIdStr, 10, 32)

	stream, _ := client.ChatStream(context.Background())
	go func() {
		for {
			msg, _ := stream.Recv()
			if msg != nil {
				fmt.Printf("[%s] %s\n", msg.GetName(), msg.GetText())
			}
		}
	}()

	stream.Send(&proto.Message{Text: "register", Name: name, RoomId: uint32(roomId)})

	for {
		text, _ := reader.ReadString('\n')
		textClean := strings.Trim(text, "\r\n")
		stream.Send(&proto.Message{Text: textClean, Name: name, RoomId: uint32(roomId)})
	}
}
