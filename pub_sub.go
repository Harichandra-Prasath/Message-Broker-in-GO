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
	r.HandleFunc("/sub/{section}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		section := vars["section"]
		if _, ok := s.Sections[section]; !ok {
			w.Write([]byte("Section not found...."))
			return
		} else {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				w.Write([]byte("Internal Error"))
				log.Fatal(err)

			}
			slog.Info("Websocket connection Made")
			conn.WriteMessage(websocket.BinaryMessage, []byte("You are connected"))
			s.Peerch <- Peer{
				Section: section,
				Conn:    conn,
			}

		}

	}).Methods("GET")

	err := http.ListenAndServe(s.Config.ConsumerListenAddr, r)
	return err
}

func (s *Server) NewSection(section string) {
	if _, ok := s.Sections[section]; !ok {
		s.Sections[section] = s.Config.StoreFunc()
		slog.Info("New Section in memory created")
	}
}

func (s *Server) NewPeer(p *Peer) error {
	// check for the section
	if _, ok := s.Sections[p.Section]; !ok {
		return fmt.Errorf("%s Section not found", p.Section)
	}
	// add the current connection to the submap
	s.Submap[p.Section] = append(s.Submap[p.Section], p.Conn)
	slog.Info("Peer added to the section")
	return nil
}

func (s *Server) publish(data *Post) {
	for _, conn := range s.Submap[data.section] {
		conn.WriteMessage(websocket.BinaryMessage, data.Data)
	}
}
