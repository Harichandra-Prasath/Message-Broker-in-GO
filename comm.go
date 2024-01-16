// defines all the structs and interfaces used for communication (transport)

package main

import (
	"encoding/json"
	"net"

	"github.com/gorilla/websocket"
)

// common inteface to support multiple protocols
type peer interface {
	Read_data(interface{}) error
	Write_data(interface{}) error
	Close() error
}

// Struct extending peer interface with managing peer sections offsets
type Peer struct {
	Conn          peer
	SectionOffset map[string]int //section to offset
}

// definition for a websocket peer
type websocket_peer struct {
	Conn *websocket.Conn
}

func (w *websocket_peer) Read_data(v interface{}) error {
	return w.Conn.ReadJSON(v)
}

func (w *websocket_peer) Write_data(v interface{}) error {
	return w.Conn.WriteJSON(v)
}

func (w *websocket_peer) Close() error {
	return w.Conn.Close()
}

// definition for a tcp peer

type tcp_peer struct {
	Conn net.Conn
}

func (t *tcp_peer) Read_data(v interface{}) error {
	buffer := make([]byte, 2048)
	n, err := t.Conn.Read(buffer)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buffer[:n], v)
	if err != nil {
		return err
	}
	return nil
}

func (t *tcp_peer) Write_data(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = t.Conn.Write(data)
	if err != nil {
		return err
	}
	return nil

}

func (t *tcp_peer) Close() error {
	return t.Conn.Close()
}

// Publishing Struct
type Post struct {
	Data    []byte
	section string
}

// Response Message to the consumer/subsriber
type Message struct {
	Status  string `json:"Status"`
	Section string `json:"Section"`
	Data    []byte `json:"Data"`
}

// Inbound Request Message from the consumer/subscriber
type Request struct {
	Reason   string   `json:"Reason"`
	Sections []string `json:"Sections"`
}
