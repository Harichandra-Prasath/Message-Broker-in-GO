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
	Sections    map[string]Storer
	Subscribers map[string][]*Peer
	Postch      chan Post
	Peerch      chan Peer
}

func Newserver(cfg *Config) *Server {
	return &Server{
		Config:      cfg,
		Sections:    make(map[string]Storer),
		Subscribers: make(map[string][]*Peer),
		Postch:      make(chan Post),
		Peerch:      make(chan Peer),
	}
}

func (s *Server) listen() {
	for {
		select {
		case data := <-s.Postch:
			s.NewSection(data.section)
			s.Sections[data.section].Push(data.Data)
			for _, peer := range s.Subscribers[data.section] {
				peer.Conn.Write_data(Message{
					Status:  "Updates",
					Section: data.section,
					Data:    []byte("New Post published...Pull to see the latest post"),
				})
			}
		case peer := <-s.Peerch:
			if len(peer.SectionOffset) == 0 {
				peer.Conn.Write_data(Message{
					Status:  "Success",
					Section: "",
					Data:    []byte("You are Connected...Subscribe to topics to Pull messages"),
				})
			} else {
				go peerlisten(&peer, s)
			}

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
