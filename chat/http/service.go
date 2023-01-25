package service

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"com.adoublef.wss/chat"
	repo "com.adoublef.wss/chat/sqlite"
	websocket "com.adoublef.wss/gobwas"
)

var _ http.Handler = (*service)(nil)

type service struct {
	r repo.Repo
	m chi.Router

	broker *websocket.Broker
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService(r repo.Repo) http.Handler {
	s := &service{
		m:      chi.NewMux(),
		r:      r,
		broker: websocket.NewBroker(),
	}

	s.routes()
	return s
}

func (s *service) routes() {
	s.m.Post("/", s.handleCreateChat())
	s.m.Get("/", s.handleListChats())
	s.m.Get("/{id}", s.handleChatInfo())
	s.m.Get("/{id}/ws", s.handleP2PConn())
}

func (s *service) handleP2PConn() http.HandlerFunc {
	parseParam := func(r *http.Request) (uuid.UUID, error) {
		return uuid.Parse(chi.URLParam(r, "id"))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := parseParam(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = s.r.Find(r.Context(), uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		cli, ok := s.broker.Find(uid.String())
		if !ok {
			cli = websocket.NewClient()
			s.broker.Add(uid.String(), cli)
		}

		cli.ServeHTTP(w, r)
	}
}

func (s *service) handleCreateChat() http.HandlerFunc {
	type response struct {
		Location string `json:"location"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		chat := chat.NewChat()

		if err := s.r.Create(r.Context(), chat); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, r, &response{
			Location: "http://localhost:8080/chats/" + chat.ID.String(),
		}, http.StatusCreated)
	}
}

func (s *service) handleChatInfo() http.HandlerFunc {
	parseParam := func(r *http.Request) (uuid.UUID, error) {
		return uuid.Parse(chi.URLParam(r, "id"))
	}

	type response struct {
		Chat *chat.Chat `json:"chat"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := parseParam(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		chat, err := s.r.Find(r.Context(), uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		respond(w, r, &response{
			Chat: chat,
		}, http.StatusOK)
	}
}

func (s *service) handleListChats() http.HandlerFunc {
	type response struct {
		Length    int          `json:"length"`
		ChatRooms []*chat.Chat `json:"chatRooms"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		cs, err := s.r.FindMany(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, r, &response{
			Length:    len(cs),
			ChatRooms: cs,
		}, http.StatusOK)
	}
}

func respond(w http.ResponseWriter, r *http.Request, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, "Could not encode in json", status)
		}
	}
}
