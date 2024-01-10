package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Request struct {
	Reason   string   `json:"Reason"`
	Sections []string `json:"Sections"`
}

type Message struct {
	Section string `json:"Topic"`
	Data    []byte `json:"Data"`
}

func main() {
	request1 := Request{
		Reason:   "Subscribe",
		Sections: []string{"foo", "bar"},
	}
	request2 := Request{
		Reason:   "Pull",
		Sections: []string{"foo", "bar"},
	}
	conn1, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:4000/sub/", nil)

	err := conn1.WriteJSON(request1)
	if err != nil {
		fmt.Print(err)
	}
	conn1.WriteJSON(request2)
	for {

		var message Message
		err := conn1.ReadJSON(&message)
		if err != nil {
			fmt.Print(err)
		}
		fmt.Printf("Data: %s recieved from server on Section: %s\n", string(message.Data), message.Section)

	}

}
