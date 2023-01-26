package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	thread "com.adoublef.wss/internal/communications"
	repo "com.adoublef.wss/internal/communications/sql/postgres"
)

var _ http.Handler = (*service)(nil)

type service struct {
	r repo.ThreadRepo
	m chi.Router

	br thread.Broker
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService(r repo.ThreadRepo) http.Handler {
	s := &service{
		m:  chi.NewMux(),
		r:  r,
		br: thread.NewBroker(),
	}

	s.routes()
	return s
}

func (s *service) routes() {
	s.m.Post("/", s.handleCreateChat())
	s.m.Get("/", s.handleListChats())

	s.m.With(threadIDMiddleware).Route("/{id}", func(r chi.Router) {
		r.Get("/", s.handleChatInfo())
		r.Delete("/", s.handleDeleteChat())
		r.Get("/ws", s.handleP2PConn())
	})
}

type contextKey string

func (k contextKey) String() string {
	return "chat context key " + string(k)
}

const (
	chatIDKey     contextKey = "chatID"
	apiVersionKey contextKey = "apiVersion"
)

func threadIDMiddleware(hf http.Handler) http.Handler {
	parseID := func(r *http.Request) (uuid.UUID, error) {
		return uuid.Parse(chi.URLParam(r, "id"))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), chatIDKey, id)
		hf.ServeHTTP(w, r.WithContext(ctx))
	})
}

func threadIDFromRequest(r *http.Request) (uuid.UUID, error) {
	return threadIDFromContext(r.Context())
}

func threadIDFromContext(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(chatIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("chat id not found")
	}

	return id, nil
}

func (s *service) handleP2PConn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, _ := threadIDFromRequest(r)
		// NOTE lookup cache first, then db
		thread, err := s.r.Find(r.Context(), uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// NOTE this may be risky
		thread, _ = s.br.LoadOrStore(uid.String(), thread)
		thread.Client().ServeHTTP(w, r)
	}
}

func (s *service) handleDeleteChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, _ := threadIDFromRequest(r)

		if err := s.r.Delete(r.Context(), uid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// NOTE async deletion is ok
		go s.br.Delete(uid.String())

		s.respond(w, r, nil, http.StatusNoContent)
	}
}

func (s *service) handleCreateChat() http.HandlerFunc {
	type response struct {
		Location string `json:"location"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// NOTE this would be created from user input
		// such as title, description and capacity
		thread := thread.NewThread()

		if err := s.r.Create(r.Context(), thread); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.br.Store(thread.ID.String(), thread)

		s.respond(w, r, &response{
			// TODO use the real host + path
			Location: "http://localhost:8080/" + thread.ID.String(),
		}, http.StatusCreated)
	}
}

func (s *service) handleChatInfo() http.HandlerFunc {
	type response struct {
		Chat *thread.Thread `json:"chat"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid, _ := threadIDFromRequest(r)

		chat, err := s.r.Find(r.Context(), uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		s.respond(w, r, &response{
			Chat: chat,
		}, http.StatusOK)
	}
}

func (s *service) handleListChats() http.HandlerFunc {
	type response struct {
		Length    int              `json:"length"`
		ChatRooms []*thread.Thread `json:"chatRooms"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		cs, err := s.r.FindMany(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.respond(w, r, &response{
			Length:    len(cs),
			ChatRooms: cs,
		}, http.StatusOK)
	}
}

func (s *service) respond(w http.ResponseWriter, r *http.Request, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, "Could not encode in json", status)
		}
	}
}
