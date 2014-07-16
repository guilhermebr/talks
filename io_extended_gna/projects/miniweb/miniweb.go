package main

import (
	"html/template"
	"net/http"
)

var templates *template.Template

func handler(w http.ResponseWriter, r *http.Request) {

	var IndexVars struct {
		Title string
	}

	IndexVars.Title = "Google IO Extended 2014"
	err := templates.ExecuteTemplate(w, "index.html", IndexVars)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	templates = template.Must(template.ParseGlob("views/*.html"))

	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("static/img"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("static/js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
