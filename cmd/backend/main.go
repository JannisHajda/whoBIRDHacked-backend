package main

import (
	"flag"
	"github.com/JannisHajda/whoBIRDHacked-backend/internal/api"
	"github.com/JannisHajda/whoBIRDHacked-backend/internal/dashboard"
	"github.com/JannisHajda/whoBIRDHacked-backend/internal/ws"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP service address")
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/ws", ws.Handler)
	http.HandleFunc("/dashboard", dashboard.DashboardHandler)
	http.HandleFunc("/api/exec", api.Handler)

	log.Println("Server started at", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
