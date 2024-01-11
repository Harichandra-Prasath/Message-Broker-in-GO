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
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

type Message struct {
	Status  string `json:"Status"`
	Section string `json:"Section"`
	Data    []byte `json:"Data"`
}

type Request struct {
	Reason   string   `json:"Reason"`
	Sections []string `json:"Sections"`
}

func StartProducer(s *Server, ro *mux.Router) error {
	ro.HandleFunc("/pub/{section}", func(w http.ResponseWriter, r *http.Request) {
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
	err := http.ListenAndServe(s.Config.ProduceListenAddr, ro)
	return err
}

func StartConsumer(s *Server, ro *mux.Router) error {
	ro.HandleFunc("/sub/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			w.Write([]byte("Error in upgrading socket"))
			log.Panic(err)
			return
		}
		slog.Info("Websocket connection Made")
		conn.WriteJSON(Message{
			Status:  "Success",
			Section: "",
			Data:    []byte("You are Connected"),
		})
		s.Peerch <- Peer{
			Conn:          conn,
			SectionOffset: make(map[string]int),
		}

	}).Methods("GET")

	err := http.ListenAndServe(s.Config.ConsumerListenAddr, ro)
	return err
}

func peerlisten(p *Peer, s *Server) {
	var request Request
	for {
		err := p.Conn.ReadJSON(&request)
		if err != nil {
			fmt.Println(request.Sections)
			slog.Error(err.Error())
		} else {
			go Process(request, p, s)
		}

	}
}

func Process(r Request, p *Peer, s *Server) {
	if r.Reason == "Subscribe" {
		go s.AddPeertoSection(p, r)
	}
	if r.Reason == "Pull" {
		go s.PushtoPeer(p, r)
	}
}

func (s *Server) PushtoPeer(p *Peer, r Request) {
	if len(r.Sections) == 0 {
		slog.Info("No sections provided")
		p.Conn.WriteJSON(Message{
			Status:  "Error",
			Section: "",
			Data:    []byte("NO section Provided to Pull"),
		})
		return
	}
	for _, section := range r.Sections {
		if _, ok := p.SectionOffset[section]; !ok {
			p.Conn.WriteJSON(Message{
				Status:  "Error",
				Section: section,
				Data:    []byte("Not subscribed to the section"),
			})
			continue
		}
		offset := p.SectionOffset[section]
		size := len(s.Sections[section].(*Store).data)
		for i := offset; i < size; i++ {
			data, err := s.Sections[section].Fetch(i)
			if err != nil {
				slog.Info(err.Error())
				p.Conn.WriteJSON(Message{
					Status:  "Error",
					Section: section,
					Data:    []byte(err.Error()),
				})
				continue
			}
			p.Conn.WriteJSON(Message{
				Status:  "Success",
				Section: section,
				Data:    data,
			})
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

func (s *Server) AddPeertoSection(p *Peer, r Request) {
	if len(r.Sections) == 0 {
		slog.Info("No sections provided")
		p.Conn.WriteJSON(Message{
			Status:  "Error",
			Section: "",
			Data:    []byte("NO section Provided to Subscribe"),
		})
		return
	}
	// check for the section
	for _, section := range r.Sections {
		if _, ok := s.Sections[section]; !ok {
			slog.Info("Section not found", "section", section)
			p.Conn.WriteJSON(Message{
				Status:  "Error",
				Section: section,
				Data:    []byte("Section not found or not yet published"),
			})
		} else {
			p.SectionOffset[section] = 0
			slog.Info("Peer added to the section", "section", section)
			p.Conn.WriteJSON(Message{
				Status:  "Success",
				Section: section,
				Data:    []byte("You are subscibed to the section"),
			})
		}

	}
}
