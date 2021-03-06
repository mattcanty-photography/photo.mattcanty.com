package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/apex/gateway/v2"
	"github.com/go-chi/chi"
)

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		log.Fatal(http.ListenAndServe(":3000", routing("")))
	} else {
		log.Fatal(gateway.ListenAndServe(":3000", routing("/dev-3918bed")))
	}
}

func routing(apiStage string) http.Handler {

	r := chi.NewRouter()

	r.HandleFunc(fmt.Sprintf("%s/", apiStage), getHome)
	r.HandleFunc(fmt.Sprintf("%s/portfolio/{id}", apiStage), getPortfolio)
	r.HandleFunc(fmt.Sprintf("%s/album/{id}", apiStage), getAlbum)
	r.HandleFunc(fmt.Sprintf("%s/photo/{id}", apiStage), getPhoto)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Not found: [%s]", r.RequestURI)))
	})

	return r
}

func getHome(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("root.html", "home.html")
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, nil)

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprint(w, b.String())
}

func getPortfolio(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("root.html", "portfolio.html")
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, nil)

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprint(w, b.String())
}

func getAlbum(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("root.html", "album.html")
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, nil)

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprint(w, b.String())
}

func getPhoto(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("root.html", "photo.html")
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, nil)

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprint(w, b.String())
}
