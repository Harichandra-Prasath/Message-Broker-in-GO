package main

import (
	"github.com/gorilla/mux"
)

type Config struct {
	ProduceListenAddr     string
	ConsumerListenAddr    string
	ConsumerListenAddrTCP string
	StoreFunc             ProduceFunc
}

type Server struct {
	*Config
	Sections map[string]Storer
	Postch   chan Post
	Peerch   chan Peer
}

func Newserver(cfg *Config) *Server {
	return &Server{
		Config:   cfg,
		Sections: make(map[string]Storer),
		Postch:   make(chan Post),
		Peerch:   make(chan Peer),
	}
}

func (s *Server) listen() {
	for {
		select {
		case data := <-s.Postch:
			s.NewSection(data.section)
			s.Sections[data.section].Push(data.Data)
		case peer := <-s.Peerch:
			go peerlisten(&peer, s)
		}

	}
}

func (s *Server) Serve() error {

	produce := mux.NewRouter()
	go StartProducer(s, produce)

	consume := mux.NewRouter()
	go StartConsumer(s, consume)

	go StartConsumertcp(s)
	s.listen()
	return nil

}
