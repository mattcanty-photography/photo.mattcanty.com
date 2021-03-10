package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func htmlRespond(w http.ResponseWriter, viewTemplateName string, data interface{}) {
	t, err := template.ParseFiles("root.html", viewTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, data)

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprint(w, b.String())
}
