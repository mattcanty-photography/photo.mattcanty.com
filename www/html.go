package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type breadcrumbs struct {
	Portfolio string
	Album     string
	Photo     string
}

func htmlRespond(w http.ResponseWriter, viewTemplateName string, data interface{}) {
	t, err := template.ParseFiles("root.html", "breadcrumbs.html", viewTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, data)

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprint(w, b.String())
}
