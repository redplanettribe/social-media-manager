package server

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/redplanettribe/social-media-manager/internal/infrastructure/config"
)

type Server interface {
	Serve()
}

func NewHttpServer(cfg *config.Config, router http.Handler) Server {
	if cfg.App.Env == "development" {
		log.Println("Development mode")
		return &HttpServer{
			cfg:    cfg,
			router: router,
		}
	}
	return &HttpsServer{
		cfg:    cfg,
		router: router,
	}
}

type HttpsServer struct {
	cfg    *config.Config
	router http.Handler
}

func (s *HttpsServer) Serve() {

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

	server := &http.Server{
		Addr:      ":" + s.cfg.App.Port,
		Handler:   s.router,
		TLSConfig: tlsConfig,
	}

	log.Printf("Server is running on port %s", s.cfg.App.Port)
	if err := server.ListenAndServeTLS(s.cfg.SSL.CertPath, s.cfg.SSL.KeyPath); err != nil {
		log.Fatal(err)
	}
}

type HttpServer struct {
	cfg    *config.Config
	router http.Handler
}

func (s *HttpServer) Serve() {
	server := &http.Server{
		Addr:                         ":" + s.cfg.App.Port,
		Handler:                      s.router,
		DisableGeneralOptionsHandler: true,
	}

	log.Printf("Dev Server is running on port %s", s.cfg.App.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
