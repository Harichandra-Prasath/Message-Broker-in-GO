package main

import (
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Post struct {
	Data    []byte
	section string
}

type Peer struct {
	Section string
	Conn    *websocket.Conn
}

type Config struct {
	ProduceListenAddr  string
	ConsumerListenAddr string
	StoreFunc          ProduceFunc
}

type Server struct {
	*Config
	Sections map[string]Storer
	Submap   map[string][]*websocket.Conn
	Postch   chan Post
	Peerch   chan Peer
}

func Newserver(cfg *Config) *Server {
	return &Server{
		Config:   cfg,
		Sections: make(map[string]Storer),
		Postch:   make(chan Post),
		Peerch:   make(chan Peer),
		Submap:   make(map[string][]*websocket.Conn),
	}
}

func (s *Server) listen() {
	for {
		select {
		case data := <-s.Postch:
			s.NewSection(data.section)
			s.StoreFunc().Append(data.Data)
			s.publish(&data)
		case peer := <-s.Peerch:
			err := s.NewPeer(&peer)
			if err != nil {
				slog.Info("Error", err)
			}
		}

	}
}

func (s *Server) Serve() error {

	produce := mux.NewRouter()
	go StartProducer(s, produce)

	consume := mux.NewRouter()
	go StartConsumer(s, consume)
	s.listen()
	return nil

}
