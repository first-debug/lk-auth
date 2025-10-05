package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"

	sl "lk-auth/internal/libs/logger"
	"lk-auth/internal/server/middleware"
	"lk-auth/internal/server/schemas"
	"lk-auth/internal/service/auth"

	"github.com/gorilla/csrf"
)

type Server struct {
	ctx            context.Context
	router         *http.ServeMux
	auth           auth.AuthService
	log            *slog.Logger
	isShuttingDown *atomic.Bool
	server         http.Server
}

type ErrMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewServer(ctx context.Context, auth auth.AuthService, log *slog.Logger, isShuttingDown *atomic.Bool) *Server {
	s := &Server{
		ctx:            ctx,
		router:         http.NewServeMux(),
		auth:           auth,
		log:            log,
		isShuttingDown: isShuttingDown,
	}

	s.router.HandleFunc("GET /ping",
		middleware.Chain(s.handlePing, middleware.Logging(log)),
	)
	s.router.HandleFunc("POST /signin",
		middleware.Chain(s.handleSignin, middleware.Logging(log)),
	)
	s.router.HandleFunc("POST /login",
		middleware.Chain(s.handleLogin, middleware.Logging(log)),
	)
	s.router.HandleFunc("POST /refresh",
		middleware.Chain(s.handleRefresh, middleware.Logging(log)),
	)
	s.router.HandleFunc("POST /logout",
		middleware.Chain(s.handleLogout, middleware.Logging(log)),
	)
	// TODO: может нужно переимновать в /validate
	s.router.HandleFunc("POST /checktoken",
		middleware.Chain(s.handleCheckToken, middleware.Logging(log)),
	)
	// TODO: добавить в OAPI спецификацию
	s.router.HandleFunc("GET /healthz", s.handleHealthz)

	return s
}

func (s *Server) Start(env, addr string) error {
	if env != "prod" {
		csrf.Secure(false)
		s.log.Debug("not a prod")
	}
	// csrfProt := csrf.Protect([]byte("32-byte-long-auth-key"))
	s.server = http.Server{
		Addr:    addr,
		Handler: s.router,
		BaseContext: func(_ net.Listener) context.Context {
			return s.ctx
		},
	}

	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.log.Error("Server failed to start", sl.Err(err))
		return err
	}
	return nil
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Pong")
}

func (s *Server) handleSignin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	signinData := schemas.SigninData{}
	err := json.NewDecoder(r.Body).Decode(&signinData)
	if err != nil {
		s.log.Debug("/signin", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  err.Error(),
			},
		)
		return
	}
	s.log.Debug("/signin", "Email", signinData.Email, "Password", signinData.Password)
	err = s.auth.Signin(signinData.Email, signinData.Password, signinData.Role)
	if err != nil {
		s.log.Debug("/signin", sl.Err(err))
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
	fmt.Fprintf(w, "Successful registration with email: %s\n", signinData.Email)
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
	s.log.Debug("/login", "Email", loginData.Email, "Password", loginData.Password)
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

	w.WriteHeader(http.StatusOK)
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
	res, err := s.auth.ValidateToken(token.AccessToken)
	if err != nil {
		s.log.Warn("token validation error", sl.Err(err))
	}
	if !res {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			ErrMsg{
				Code: 400,
				Msg:  "token is invalid",
			},
		)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if s.isShuttingDown.Load() {
		http.Error(w, "Shutting down", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "OK")
}

func (s *Server) ShutDown(shutDownCtx context.Context) error {
	return s.server.Shutdown(shutDownCtx)
}
