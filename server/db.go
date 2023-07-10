package main

import (
	"database/sql"
	"fmt"
	"log"

	proto "talktunnel"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "admin"
	dbname   = "chatDB"
)

var conn *sql.DB

func startDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			text TEXT NOT NULL,
			name TEXT NOT NULL,
			room_id INT NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}

	conn = db
}

func addMessage(msg *proto.Message) {
	_, err := conn.Exec("INSERT INTO messages (text, name, room_id) VALUES ($1, $2, $3)", msg.Text, msg.Name, msg.RoomId)
	if err != nil {
		log.Fatal(err)
	}
}

func getMessages(roomId uint32) []*proto.Message {
	rows, err := conn.Query("SELECT text, name, room_id FROM messages WHERE room_id=$1", roomId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	messages := make([]*proto.Message, 0)
	for rows.Next() {
		m := &proto.Message{}
		err = rows.Scan(&m.Text, &m.Name, &m.RoomId)
		if err != nil {
			panic(err)
		}
		messages = append(messages, m)
	}
	return messages
}
