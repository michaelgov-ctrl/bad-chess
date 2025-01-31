package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	serverURL := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}

	/*
		//type port
		buf := bufio.NewReader(os.Stdin)
		input, err := buf.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		header := http.Header{}
		header.Add("origin", fmt.Sprintf("ws://localhost:%s", input))
	*/
	log.Printf("connecting to %s...", serverURL.String())
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil) //header)
	if err != nil {
		log.Fatalf("failed to connect to ws server: %v", err)
	}
	defer conn.Close()

	go func() {
		for {
			_, reply, err := conn.ReadMessage()
			if err != nil {
				log.Fatalf("failed to read message: %v", err)
			}
			log.Printf("received: %s\n", reply)
		}
	}()

	buf := bufio.NewReader(os.Stdin)
	for {
		time.Sleep(time.Second * 1)

		fmt.Println("make your move")
		input, err := buf.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}

		//message := fmt.Sprintf(`{"type":"make_move","payload":%s}`, input)
		// {"type":"join_match","payload":{"time_control":"20m"}}
		err = conn.WriteMessage(websocket.TextMessage, []byte(input))
		if err != nil {
			log.Fatalf("failed to send message: %v", err)
		}
		log.Printf("sent: %s\n", input)

	}

}
