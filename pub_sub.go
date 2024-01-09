package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{}
)

type Message struct {
	Reason   string   `json:"Reason"`
	Sections []string `json:"Sections"`
}

func StartProducer(s *Server, r *mux.Router) error {
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
	err := http.ListenAndServe(s.Config.ProduceListenAddr, r)
	return err
}

func StartConsumer(s *Server, r *mux.Router) error {
	r.HandleFunc("/sub/", func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.Write([]byte("Internal Error"))
			log.Panic(err)
			return
		}

		slog.Info("Websocket connection Made")
		conn.WriteMessage(websocket.BinaryMessage, []byte("You are connected"))
		s.Peerch <- Peer{
			Conn:          conn,
			SectionOffset: make(map[string]int),
		}
	}).Methods("GET")

	err := http.ListenAndServe(s.Config.ConsumerListenAddr, r)
	return err
}

func peerlisten(p *Peer, s *Server) {
	var message Message
	for {
		err := p.Conn.ReadJSON(&message)
		if err != nil {
			fmt.Println(message.Sections)
			slog.Error(err.Error())
		} else {
			Process(message, p, s)
		}

	}
}

func Process(m Message, p *Peer, s *Server) {
	if m.Reason == "Subscribe" {
		s.AddPeertoSection(p, m)
	}
	if m.Reason == "Pull" {
		s.PushtoPeer(p, m)
	}
}

func (s *Server) PushtoPeer(p *Peer, m Message) {
	if len(m.Sections) == 0 {
		slog.Info("No sections provided")
		p.Conn.WriteMessage(websocket.BinaryMessage, []byte("Give a section"))
		return
	}
	for _, section := range m.Sections {
		if _, ok := p.SectionOffset[section]; !ok {
			p.Conn.WriteMessage(websocket.BinaryMessage, []byte("You are not subscribed to this section"))
			continue
		}
		offset := p.SectionOffset[section]
		size := len(s.Sections[section].(*Store).data)
		fmt.Println(size)
		for i := offset; i < size; i++ {
			data, err := s.Sections[section].Fetch(i)
			if err != nil {
				slog.Info(err.Error())
				p.Conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
				continue
			}
			p.Conn.WriteMessage(websocket.BinaryMessage, []byte(data))
			p.SectionOffset[section] += 1

		}

	}

}

func (s *Server) NewSection(section string) {
	if _, ok := s.Sections[section]; !ok {
		s.Sections[section] = s.Config.StoreFunc()
		slog.Info("New Section in memory created")
	}
}

func (s *Server) AddPeertoSection(p *Peer, m Message) {
	// check for the section
	for _, section := range m.Sections {
		if _, ok := s.Sections[section]; !ok {
			slog.Info("Section not found", "section", section)
		} else {
			p.SectionOffset[section] = 0
			slog.Info("Peer added to the section", "section", section)
		}

	}
}
