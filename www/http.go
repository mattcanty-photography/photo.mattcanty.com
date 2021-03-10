package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routing() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.Compress(9, "text/html"))

	r.HandleFunc("/", getHome)
	r.HandleFunc("/portfolio/{portfolioID}", getPortfolio)
	r.HandleFunc("/portfolio/{portfolioID}/album/{albumID}", getAlbum)
	r.HandleFunc("/portfolio/{portfolioID}/album/{albumID}/photo/{photoID}", getPhoto)

	r.HandleFunc("/{apiStage}/", getHome)
	r.HandleFunc("/{apiStage}/portfolio/{portfolioID}", getPortfolio)
	r.HandleFunc("/{apiStage}/portfolio/{portfolioID}/album/{albumID}", getAlbum)
	r.HandleFunc("/{apiStage}/portfolio/{portfolioID}/album/{albumID}/photo/{photoID}", getPhoto)

	r.NotFound(notFound)

	return r
}
