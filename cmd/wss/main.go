package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strconv"

	srv "com.adoublef.wss/internal/communications/http"
	repo "com.adoublef.wss/internal/communications/sql/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

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
var connStr = flag.String("f", "postgres://postgres:postgres@localhost:5432/", "create a new sqlite3 database")

func init() {
	flag.Parse()
}

func run() error {
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, *connStr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		return err
	}

	chatSrv := newChatService(conn)

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

func newChatService(conn *pgxpool.Pool) http.Handler {
	repo := repo.NewChatRepo(conn)
	return srv.NewService(repo)
}
