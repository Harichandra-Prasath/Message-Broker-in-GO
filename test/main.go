package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

type Request struct {
	Reason   string   `json:"Reason"`
	Sections []string `json:"Sections"`
}

type Message struct {
	Status  string `json:"Status"`
	Section string `json:"Section"`
	Data    []byte `json:"Data"`
}

func main() {
	tcpserver, _ := net.ResolveTCPAddr("tcp", "localhost:5000")
	_, err := http.Post("http://127.0.0.1:3000/pub/foo", "application/octet-stream", bytes.NewReader([]byte("foo")))
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < 1000; i++ {
		go func() {
			conn, err := net.DialTCP("tcp", nil, tcpserver)
			if err != nil {
				fmt.Print(err)
			}
			data, _ := json.Marshal(Request{
				Reason:   "Subscribe",
				Sections: []string{"foo"},
			})
			_, err = conn.Write(data)
			if err != nil {
				fmt.Print("Error")
			}
			if err != nil {
				println("Write data failed:", err.Error())
				os.Exit(1)
			}
			for {
				recieved := make([]byte, 1024)
				_, err = conn.Read(recieved)
				if err != nil {
					fmt.Println("failed")
				}
				fmt.Println("Recieved Message")
			}

		}()
		time.Sleep(500 * time.Millisecond)
	}
	for {
		continue
	}
}
