package dashboard

import (
	"github.com/JannisHajda/whoBIRDHacked-backend/internal/ws"
	"html/template"
	"net/http"
	"path/filepath"
)

var templates = template.Must(template.ParseGlob(filepath.Join("internal", "dashboard", "templates", "*.html")))

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	clients := ws.GetClients()
	data := struct {
		Clients map[string]ws.Client
	}{
		Clients: clients,
	}

	renderTemplate(w, "dashboard.html", data)
}
