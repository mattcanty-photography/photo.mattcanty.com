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
		Breadcrumbs breadcrumbs
		portfolioID string
		Albums      []album
	}{
		Breadcrumbs: breadcrumbs{
			Portfolio: portfolioID,
		},
		portfolioID: portfolioID,
		Albums:      getAlbums(portfolioID),
	}

	htmlRespond(w, "portfolio.html", viewData)
}

func getAlbum(w http.ResponseWriter, r *http.Request) {
	portfolioID := chi.URLParam(r, "portfolioID")
	albumID := chi.URLParam(r, "albumID")

	viewData := struct {
		Breadcrumbs breadcrumbs
		Photos      []photo
	}{
		Breadcrumbs: breadcrumbs{
			Portfolio: portfolioID,
			Album:     albumID,
		},
		Photos: getPhotos(portfolioID, albumID),
	}

	htmlRespond(w, "album.html", viewData)
}

func getPhoto(w http.ResponseWriter, r *http.Request) {
	portfolioID := chi.URLParam(r, "portfolioID")
	albumID := chi.URLParam(r, "albumID")
	photoID := chi.URLParam(r, "photoID")

	viewData := struct {
		Breadcrumbs breadcrumbs
		PortfolioID string
		AlbumID     string
		PhotoID     string
	}{
		Breadcrumbs: breadcrumbs{
			Portfolio: portfolioID,
			Album:     albumID,
			Photo:     photoID,
		},
		PortfolioID: portfolioID,
		AlbumID:     albumID,
		PhotoID:     photoID,
	}

	htmlRespond(w, "photo.html", viewData)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	htmlRespond(w, "404.html", nil)
}
