package main

import (
	"log"
	"net"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/bashar-515/project-banana/gen/go/app/v1/appv1connect"
	"github.com/bashar-515/project-banana/src/backend/lib/server"
)

func main() {
	env, err := godotenv.Read()
	if err != nil {
		log.Fatalf("error reading .env: %v", err)
	}

	port, ok := env["PB_SERVER_PORT"]
	if !ok {
		log.Fatalf("server port not found in .env. Killing server")
	}

	host, ok := env["PB_SERVER_HOST"]
	if !ok {
		log.Fatalf("server port not found in .env. Killing server")
	}

	s := server.NewServer()
	mux := http.NewServeMux()

	mux.Handle(appv1connect.NewAppServiceHandler(s))
	mux.HandleFunc("/ws", s.HandleWebSocket)

	addr := net.JoinHostPort(host, port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error listening at address %s: %v", addr, err)
	}

	log.Printf("listening at address %s", listener.Addr())
	log.Fatal(
		http.Serve(
			listener,
			cors.New(cors.Options{
				AllowedOrigins: []string{"https://app.project-banana.com"},
				AllowedHeaders: []string{"*"},
			}).Handler(mux),
		),
	)
}
