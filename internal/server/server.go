package server

import (
	sl "auth-service/internal/libs/logger"
	"auth-service/internal/server/middleware"
	"auth-service/internal/server/schemas"
	"auth-service/internal/services/auth"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/csrf"
)

type Server struct {
	router *http.ServeMux
	auth   auth.AuthService
	log    *slog.Logger
}

// TODO: обдумать формат возврата ошибок
type ErrMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewServer(auth auth.AuthService, log *slog.Logger) *Server {
	s := &Server{
		router: http.NewServeMux(),
		auth:   auth,
		log:    log,
	}

	s.router.HandleFunc("GET /ping",
		middleware.Chain(s.handlePing, middleware.Logging(log)))

	s.router.HandleFunc("POST /login",
		middleware.Chain(s.handleLogin, middleware.Logging(log)),
	)
	s.router.HandleFunc("POST /refresh",
		middleware.Chain(s.handleRefresh, middleware.Logging(log)),
	)
	s.router.HandleFunc("POST /logout",
		middleware.Chain(s.handleLogout, middleware.Logging(log)),
	)
	s.router.HandleFunc("POST /checktoken",
		middleware.Chain(s.handleCheckToken, middleware.Logging(log)),
	)
	return s
}

func (s *Server) Start(env, addr string) {
	if env != "prod" {
		csrf.Secure(false)
		s.log.Debug("not a prod")
	}
	// csrfProt := csrf.Protect([]byte("32-byte-long-auth-key"))

	http.ListenAndServe(addr, s.router)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Pong")
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	loginData := schemas.LoginData{}
	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		s.log.Debug("/login", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}
	accessToken, refreshToken, err := s.auth.Login(loginData.Email, loginData.Password)
	if err != nil {
		s.log.Debug("/login", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}

	json.NewEncoder(w).Encode(schemas.Tokens{
		Access_token:  accessToken,
		Refresh_token: refreshToken,
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	inputToken := struct {
		Token string `json:"refresh_token"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&inputToken)
	if err != nil {
		s.log.Debug("/refresh", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}

	accessToken, refreshToken, err := s.auth.Refresh(inputToken.Token)
	if err != nil {
		s.log.Debug("/refresh", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}

	json.NewEncoder(w).Encode(schemas.Tokens{
		Access_token:  accessToken,
		Refresh_token: refreshToken,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := struct {
		AccessToken string `json:"access_token"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}
	err = s.auth.Logout(token.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCheckToken(w http.ResponseWriter, r *http.Request) {
	token := struct {
		AccessToken string `json:"access_token"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}
	res, err := s.auth.ValidateToken(token.AccessToken)
	if err != nil {
		s.log.Warn("token validation error", sl.Err(err))
	}
	if !res {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
