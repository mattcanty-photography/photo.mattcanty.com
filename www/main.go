package main

import (
	"log"
	"net/http"
	"os"

	"github.com/apex/gateway/v2"
	"github.com/go-chi/chi"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		log.Fatal(http.ListenAndServe(":3000", routing()))
	} else {
		log.Fatal(gateway.ListenAndServe(":3000", routing()))
	}
}

func getHome(w http.ResponseWriter, r *http.Request) {
	htmlRespond(w, "home.html", nil)
}

func getPortfolio(w http.ResponseWriter, r *http.Request) {
	portfolioID := chi.URLParam(r, "portfolioID")

	viewData := struct {
		portfolioID string
		Albums      []album
	}{
		portfolioID: portfolioID,
		Albums:      getAlbums(portfolioID),
	}

	htmlRespond(w, "portfolio.html", viewData)
}

func getAlbum(w http.ResponseWriter, r *http.Request) {
	viewData := struct {
		Photos []photo
	}{
		Photos: getPhotos(chi.URLParam(r, "portfolioID"), chi.URLParam(r, "albumID")),
	}

	htmlRespond(w, "album.html", viewData)
}

func getPhoto(w http.ResponseWriter, r *http.Request) {
	viewData := struct {
		PortfolioID string
		AlbumID     string
		PhotoID     string
	}{
		PortfolioID: chi.URLParam(r, "portfolioID"),
		AlbumID:     chi.URLParam(r, "albumID"),
		PhotoID:     chi.URLParam(r, "photoID"),
	}

	htmlRespond(w, "photo.html", viewData)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	htmlRespond(w, "404.html", nil)
}
