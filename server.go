package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Config struct {
	ListenAddr string
	StoreFunc  ProduceFunc
}

type Server struct {
	*Config
	Sections map[string]Storer
}

func Newserver(cfg *Config) *Server {
	return &Server{
		Config:   cfg,
		Sections: make(map[string]Storer),
	}
}

func (s *Server) Serve() error {
	r := mux.NewRouter()
	err := http.ListenAndServe(s.Config.ListenAddr, r)
	return err
}

func (s *Server) NewSection(section string) {
	if _, ok := s.Sections[section]; !ok {
		s.Sections[section] = s.Config.StoreFunc()
	}

}
