package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"
)

func staticHandler() http.Handler {
	fs := http.FileServer(http.Dir("static"))
	return http.StripPrefix("/static/", fs)
}

func IndexContrl(w http.ResponseWriter, r *http.Request) {
	render(w, "index.page.gohtml")
}

func AddContrl(w http.ResponseWriter, r *http.Request) {
	render(w, "add.page.gohtml")
}

func EditContrl(w http.ResponseWriter, r *http.Request) {
	render(w, "edit.page.gohtml")
}

func FetchHandler(w http.ResponseWriter, r *http.Request) {
	// Make an HTTP GET request
	response, err := http.Get("https://example.com") // Replace with your desired URL
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// Set the content type header based on the response
	w.Header().Set("Content-Type", response.Header.Get("Content-Type"))

	// Set the status code based on the response
	w.WriteHeader(response.StatusCode)

	// Return the response body without reading it
	_, err = io.Copy(w, response.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	userName := "admin"
	password := "GOBright@2023"

	if userName == "" || password == "" {
		log.Fatal("username/password not provided")
	}

	authHandler := NewAuthHandler(userName, password)

	// http.Handle("/", authHandler.basicAuth(staticHandler()))

	tmplHandler := http.HandlerFunc(IndexContrl)
	http.Handle("/", authHandler.basicAuth(tmplHandler))

	addHandler := http.HandlerFunc(AddContrl)
	http.Handle("/add", authHandler.basicAuth(addHandler))

	editHandler := http.HandlerFunc(EditContrl)
	http.Handle("/edit", authHandler.basicAuth(editHandler))

	log.Print("Listening on :4200...")
	err := http.ListenAndServe(":4200", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func render(w http.ResponseWriter, t string) {

	partials := []string{
		"./templates/base.layout.gohtml",
		"./templates/header.partial.gohtml",
		"./templates/footer.partial.gohtml",
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("./templates/%s", t))

	for _, x := range partials {
		templateSlice = append(templateSlice, x)
	}

	tmpl, err := template.ParseFiles(templateSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type auth struct {
	userName string
	password string
}

func NewAuthHandler(userName, password string) *auth {
	return &auth{userName, password}
}

func (a *auth) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))

			expectedUsernameHash := sha256.Sum256([]byte(a.userName))
			expectedPasswordHash := sha256.Sum256([]byte(a.password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
