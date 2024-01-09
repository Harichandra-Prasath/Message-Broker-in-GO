package main

import (
	"github.com/gorilla/mux"
)

type Post struct {
	Data    []byte
	section string
}

type Config struct {
	ProduceListenAddr  string
	ConsumerListenAddr string
	StoreFunc          ProduceFunc
}

type Server struct {
	*Config
	Sections map[string]Storer
	Postch   chan Post
}

func Newserver(cfg *Config) *Server {
	return &Server{
		Config:   cfg,
		Sections: make(map[string]Storer),
		Postch:   make(chan Post),
	}
}

func (s *Server) listen() {
	for {
		data := <-s.Postch
		s.NewSection(data.section)
		s.StoreFunc().Append(data.Data)
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
