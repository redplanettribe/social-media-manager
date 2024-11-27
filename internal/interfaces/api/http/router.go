package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/handlers"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/auth"
)

func NewRouter(postHandler *handlers.PostHandler, authenticator auth.Authenticator) http.Handler {
	router := mux.NewRouter()

	authMiddleware := middlewares.AuthMiddleware(authenticator)

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware)

	apiRouter.HandleFunc("/posts", postHandler.CreatePost).Methods("POST")
	// Other routes...

	return router
}
