package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"strconv"

	srv "com.adoublef.wss/chat/http"
	repo "com.adoublef.wss/chat/sqlite"
	"github.com/go-chi/chi/v5"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed index.html
var indexHTML []byte

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

var port = flag.Int("p", 8080, "port that server will run on")
var connStr = flag.String("f", "file:wss.db", "create a new sqlite3 database")

func init() {
	flag.Parse()
}

func run() error {
	db, err := sql.Open("sqlite3", *connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	chatSrv := newChatService(db)

	rootMux := chi.NewMux()
	rootMux.HandleFunc("/*", serveIndex)

	rootMux.Mount("/chats", chatSrv)

	srv := http.Server{
		Addr:    ":" + strconv.Itoa(*port),
		Handler: rootMux,
	}

	log.Printf("Listening %s\n", srv.Addr)
	return srv.ListenAndServe()
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
}

func newChatService(db *sql.DB) http.Handler {
	chatRepo := repo.NewRepo(db)
	chatSrv := srv.NewService(chatRepo)
	return chatSrv
}
