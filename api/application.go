package api

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/JhonX2011/GOWebApplication/api/utils/logger"
	"github.com/JhonX2011/GOWebApplication/api/web"
)

const (
	_defaultWebApplicationPort = "8080"
	_defaultNetworkProtocol    = "tcp"
)

type Application struct {
	*web.Router
	Logger logger.Logger
	
	address string
}

func NewWebApplication() (*Application, error) {
	l := logger.NewLogger(logger.DefaultOSExit)

	port := os.Getenv("PORT")
	if port == "" {
		port = _defaultWebApplicationPort
	}

	address := ":" + port

	listener, err := net.Listen(_defaultNetworkProtocol, address)
	if err != nil {
		l.Fatalf("The provided port [%s] is not available: %v", address, err)
		return nil, err
	}
	l.Info("Running application | address", address)
	defer listener.Close()

	return &Application{
		Router:  web.New(),
		Logger:  l,
		address: address,
	}, nil
}

func (a *Application) Run() error {
	a.defaultRoutes()

	srv := &http.Server{
		Addr:         a.address,
		Handler:      a.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (a *Application) defaultRoutes() {
	a.Router.Get("/ping", func(w http.ResponseWriter, r *http.Request) error {
		return web.EncodeJSON(w, "pong", 200)
	})
}
