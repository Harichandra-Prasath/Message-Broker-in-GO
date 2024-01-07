package main

import (
	"log"
	"net/http"
)

type Config struct {
	ListenAddr string
}

type Server struct {
	*Config
}

func Newserver(cfg *Config) *Server {
	return &Server{
		cfg,
	}
}

func (s *Server) Serve() error {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.URL.Path)
	})
	err := http.ListenAndServe(s.Config.ListenAddr, nil)
	return err
}
