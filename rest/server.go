package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

const (
	UserKey string = "user"
)

type Server struct {
	Store  *Store
	Broker *Broker
	Mux    *http.ServeMux
	Hook   Hook
}

func NewServer(dataDir, tmplDir, staticDir string) (*Server, error) {
	store, err := NewStore(dataDir)
	if err != nil {
		return nil, err
	}

	s := &Server{Store: store, Broker: &Broker{channels: map[string]map[chan Event]struct{}{}}, Mux: http.NewServeMux(), Hook: nopHook}
	auth := func(next http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := r.PathValue("resource")
			action := map[string]string{"GET": "read", "POST": "create", "PUT": "update", "DELETE": "delete"}[r.Method]
			user, _ := s.Store.Authenticate(r)
			if resource != "" && action != "" {
				if err := s.Store.Authorize(resource, r.PathValue("id"), action, user); err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			}
			next(w, r.WithContext(context.WithValue(r.Context(), UserKey, user)))
		})
	}

	s.Mux.Handle("GET /api/{resource}/", auth(s.handleList))
	s.Mux.Handle("POST /api/{resource}/", auth(s.handleCreate))
	s.Mux.Handle("GET /api/{resource}/{id}", auth(s.handleGet))
	s.Mux.Handle("PUT /api/{resource}/{id}", auth(s.handleUpdate))
	s.Mux.Handle("DELETE /api/{resource}/{id}", auth(s.handleDelete))
	s.Mux.HandleFunc("GET /api/events/{resource}", s.handleEvents)
	s.Mux.HandleFunc("POST /api/login", s.handleLogin)
	s.Mux.HandleFunc("POST /api/logout", s.handleLogout)
	if tmplDir != "" {
		if tmpl, err := template.ParseGlob(filepath.Join(tmplDir, "*")); err == nil {
			for _, t := range tmpl.Templates() {
				if t.Name() == "index.html" {
					s.Mux.Handle("GET /", s.handleTemplate(t, "index.html"))
				}
				s.Mux.Handle(fmt.Sprintf("GET /%s", t.Name()), s.handleTemplate(tmpl, t.Name()))
			}
		} else {
			log.Fatal("Error parsing templates:", err)
		}
	}
	if staticDir != "" {
		s.Mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	}

	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.Mux.ServeHTTP(w, r) }

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	res, err := s.Store.List(r.PathValue("resource"), r.FormValue("sort_by"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(res)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var res Resource
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resource := r.PathValue("resource")
	if err := s.Hook("create", resource, r.Context().Value("user").(Resource), res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := s.Store.Create(resource, res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Broker.Publish(resource, Event{Action: "created", ID: res["_id"].(string), Data: res})
	w.Header().Set("Location", fmt.Sprintf("/api/%s/%s", resource, id))
	w.Header().Set("HX-Trigger", fmt.Sprintf("%s-changed", resource))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	res, err := s.Store.Get(r.PathValue("resource"), r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res == nil {
		http.NotFound(w, r)
		return
	}
	_ = json.NewEncoder(w).Encode(res)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	res := Resource{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resource := r.PathValue("resource")
	res["_id"] = r.PathValue("id")
	if err := s.Hook("update", resource, r.Context().Value("user").(Resource), res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.Store.Update(resource, res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Broker.Publish(resource, Event{Action: "updated", ID: res["_id"].(string), Data: res})
	w.Header().Set("HX-Trigger", fmt.Sprintf("%s-changed", resource))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	res, _ := s.Store.Get(r.PathValue("resource"), r.PathValue("id"))
	if err := s.Hook("delete", r.PathValue("resource"), r.Context().Value("user").(Resource), res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.Store.Delete(r.PathValue("resource"), r.PathValue("id")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Broker.Publish(r.PathValue("resource"), Event{Action: "deleted", Data: res})
	w.Header().Set("HX-Trigger", fmt.Sprintf("%s-changed", r.PathValue("resource")))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	username, password := r.FormValue("username"), r.FormValue("password")
	if _, err := s.Store.AuthenticateBasic(username, password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    SignSession(username),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", HttpOnly: true, MaxAge: -1})
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleTemplate(tmpl *template.Template, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := s.Store.Authenticate(r)
		data := map[string]any{
			"Store":   s.Store,
			"Request": r,
			"User":    user,
			"ID":      r.URL.Query().Get("_id"),
			"Authorize": func(resource, id, action string) bool {
				return s.Store.Authorize(resource, id, action, user) == nil
			},
		}
		if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
			log.Println("Error executing template:", name, err)
		}
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	resource := r.PathValue("resource")
	user, err := s.Store.Authenticate(r)
	if err != nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	events := make(chan Event, 10)
	s.Broker.Subscribe(resource, events)
	defer s.Broker.Unsubscribe(resource, events)
	for {
		select {
		case e := <-events:
			if e.Action == "delete" || s.Store.Authorize(resource, e.ID, "read", user) == nil {
				data, _ := json.Marshal(e.Data)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Action, data)
				flusher.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}
