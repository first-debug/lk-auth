package server

import (
	"net/http"
)

type Server struct {
	router *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		router: http.NewServeMux(),
	}

	s.router.HandleFunc("/login", s.handleLogin)
	s.router.HandleFunc("/refresh", s.handleRefresh)
	s.router.HandleFunc("/logout", s.handleLogout)
	return s
}

func (s *Server) Start() {
	http.ListenAndServe(":8080", s.router)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: вход пользователя
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("login handler not implemented"))
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	// TODO: регистрация пользователя
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("refresh handler not implemented"))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// TODO: проверка токена
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("logout token handler not implemented"))
}
