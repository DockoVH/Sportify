package server

import (
	"net/http"
	"strings"
	"log"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Server struct {
	port int
}

func (server *Server) HandlerInit(mongoClient *mongo.Client) http.Handler {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")
		fs.ServeHTTP(w, r)
	})

	mux.HandleFunc("/", handleHome)

	return mux
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	log.Print(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func NoviServer(mongoClient *mongo.Client) (*http.Server, int) {
	noviServer := Server {
		port: 8080,
	}

	return &http.Server {
		Addr: fmt.Sprintf(":%d", noviServer.port),
		Handler: noviServer.HandlerInit(mongoClient),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}, noviServer.port
}
