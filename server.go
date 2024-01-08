package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Post struct {
	Data    []byte
	section string
}

type Config struct {
	ListenAddr string
	StoreFunc  ProduceFunc
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
	go s.listen()
	r := mux.NewRouter()
	r.HandleFunc("/pub/{section}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal("Error")
		}
		s.Postch <- Post{
			section: vars["section"],
			Data:    bytes,
		}
	}).Methods("POST")
	err := http.ListenAndServe(s.Config.ListenAddr, r)
	return err
}
func (s *Server) NewSection(section string) {
	if _, ok := s.Sections[section]; !ok {
		s.Sections[section] = s.Config.StoreFunc()
		slog.Info("New Section in memory created")
	}

}
