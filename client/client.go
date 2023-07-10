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

func readInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return ""
	}
	return strings.Trim(input, "\r\n")
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer conn.Close()

	client := proto.NewChatServiceClient(conn)
	reader := bufio.NewReader(os.Stdin)

	name := readInput(reader, "username:")
	roomIdStr := readInput(reader, "roomId:")
	roomId, err := strconv.ParseUint(roomIdStr, 10, 32)
	if err != nil {
		fmt.Println("Failed to parse room ID:", err)
		return
	}

	stream, err := client.ChatStream(context.Background())
	if err != nil {
		fmt.Println("Failed to open stream:", err)
		return
	}

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
