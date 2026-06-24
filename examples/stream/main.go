package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AtoriUzawa/cira"
	"github.com/gorilla/websocket"
)

func main() {
	server := cira.New()
	var client *cira.Conn
	server.OnConnect = func(conn *cira.Conn) {
		fmt.Println("server connect")
		client = conn
	}

	go func() { panic(server.Run(":8082")) }()

	conn, _, _ := websocket.DefaultDialer.Dial(
		"ws://localhost:8082/ws",
		nil,
	)

	time.Sleep(time.Second)

	go func() {
		client.Do(func(c *cira.Context) {
			stream := c.OpenStream("video")
			for {
				var str string
				_ = stream.Recv(&str)
				fmt.Printf("client recv str: %s\n", str)
			}
		})
	}()

	count := 1
	for {
		time.Sleep(time.Second)
		err := conn.WriteJSON(map[string]string{"type": "stream", "reply_to": "video", "data": "image_" + strconv.Itoa(count)})
		if err != nil {
			fmt.Println(err)
		}
		count++
	}
}
