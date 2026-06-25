package main

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/AtoriUzawa/cira"
)

func main() {
	server := cira.New()
	server.On("hello", func(c *cira.Context) {
		fmt.Printf("receive client message: %s\n", c.Message.Data)
		_ = c.Push("hello", "world")
	})
	server.On("upload.file", func(c *cira.Context) {
		streamID := "upload.file.123"
		c.Resp(map[string]string{"stream_id": streamID})

		stream := c.OpenStream(streamID)
		defer c.CloseStream()

		for {
			var file string
			err := stream.Recv(&file)
			if err == io.EOF {
				fmt.Println("client stream close")
				break
			}
			if err != nil {
				return
			}
			fmt.Printf("recv file: %s\n", file)
		}
		fmt.Println("transport finish")
	})

	go func() {
		panic(server.Run(":18887"))
	}()

	time.Sleep(time.Second)

	client := cira.New()
	client.On("hello", func(c *cira.Context) {
		fmt.Printf("receive server message: %s\n", c.Message.Data)
	})

	conn, err := client.Dial("ws://localhost:18887/ws")
	if err != nil {
		panic(err)
	}

	conn.Do(func(c *cira.Context) {
		_ = c.Push("hello", "hello")
	})

	conn.Do(func(c *cira.Context) {
		var resp map[string]string
		if err := c.Call("upload.file", nil, &resp); err != nil {
			return
		}
		streamID := resp["stream_id"]
		if streamID == "" {
			return
		}

		stream := c.OpenStream(streamID)
		defer c.CloseStream()

		count := 1
		for count <= 10 {
			time.Sleep(time.Second)
			err := stream.Send("file_" + strconv.Itoa(count))
			if err != nil {
				fmt.Println(err)
				return
			}
			count++
		}
	})

	select {}
}
