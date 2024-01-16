package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

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
		s.Peerch <- Peer{
			Conn: &websocket_peer{
				Conn: conn,
			},
			SectionOffset: make(map[string]int),
		}

	}).Methods("GET")

	err := http.ListenAndServe(s.Config.ConsumerListenAddr, ro)
	return err
}

func StartConsumertcp(s *Server) error {
	ln, err := net.Listen("tcp", s.ConsumerListenAddrTCP)
	if err != nil {

		log.Fatal(err)
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Info("Error", "err:", err)
			continue
		}
		slog.Info("Connection Established with", "addr", conn.RemoteAddr())
		s.Peerch <- Peer{
			Conn: &tcp_peer{
				Conn: conn,
			},
			SectionOffset: make(map[string]int),
		}
	}

}

func peerlisten(p *Peer, s *Server) {
	var request Request
	for {
		err := p.Conn.Read_data(&request)
		if err != nil {
			fmt.Print(err)
			p.Conn.Close()
			return

		} else {
			go Process(request, p, s)
		}
	}
}

func Process(r Request, p *Peer, s *Server) {
	if r.Reason == "Subscribe" {
		go s.AddPeertoSection(p, r)
	} else if r.Reason == "Pull" {
		go s.PushtoPeer(p, r)
	} else {
		p.Conn.Write_data(Message{
			Status:  "Error",
			Section: "",
			Data:    []byte("Invalid Request"),
		})
	}
}

func (s *Server) PushtoPeer(p *Peer, r Request) {
	if len(r.Sections) == 0 {
		slog.Info("No sections provided")
		p.Conn.Write_data(Message{
			Status:  "Error",
			Section: "",
			Data:    []byte("NO section Provided to Pull"),
		})
		return
	}
	for _, section := range r.Sections {
		if _, ok := s.Sections[section]; !ok {
			slog.Info("Section not found", "section", section)
			p.Conn.Write_data(Message{
				Status:  "Error",
				Section: section,
				Data:    []byte("Section not found or not yet published"),
			})
			continue
		}
		if _, ok := p.SectionOffset[section]; !ok {
			p.Conn.Write_data(Message{
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
				p.Conn.Write_data(Message{
					Status:  "Error",
					Section: section,
					Data:    []byte(err.Error()),
				})
				continue
			}
			p.Conn.Write_data(Message{
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
		p.Conn.Write_data(Message{
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
			p.Conn.Write_data(Message{
				Status:  "Error",
				Section: section,
				Data:    []byte("Section not found or not yet published"),
			})
		} else {
			s.Subscribers[section] = append(s.Subscribers[section], p)
			p.SectionOffset[section] = 0
			slog.Info("Peer added to the section", "section", section)
			p.Conn.Write_data(Message{
				Status:  "Success",
				Section: section,
				Data:    []byte("You are subscibed to the section"),
			})
		}

	}
}
