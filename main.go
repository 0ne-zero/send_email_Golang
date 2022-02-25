package main

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/gorilla/mux"
	gomail "gopkg.in/mail.v2"
)

type TemplateData struct {
	Message         string
	Method          string
	Success         bool
	ValidationError bool
}

func newRouter() *mux.Router {
	// Config routing
	r := mux.NewRouter()
	r.HandleFunc("/", indexPageGET).Methods("GET")
	r.HandleFunc("/", indexPagePOST).Methods("POST")

	// Access to static directory
	static_dir := http.Dir("./static/")
	handler := http.StripPrefix("/static/", http.FileServer(static_dir))
	r.PathPrefix("/static/").Handler(handler).Methods("GET")

	return r
}

func main() {
	fmt.Println("I'm alive")

	// Run server
	r := newRouter()
	http.ListenAndServe(":8080", r)
}

func indexPageGET(w http.ResponseWriter, r *http.Request) {
	templates, err := template.ParseFiles("./static/templates/index.gohtml")
	if err != nil {
		panic(err)
	}
	templates.ExecuteTemplate(w, "index.gohtml", TemplateData{Message: "", Success: false, Method: "GET", ValidationError: false})
}

func indexPagePOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	from := r.Form.Get("from")
	pass := r.Form.Get("pass")
	to := r.Form.Get("to")
	subject := r.Form.Get("subject")
	body := r.Form.Get("message")

	template, err := template.ParseFiles("./static/templates/index.gohtml")
	if err != nil {
		panic(err)
	}

	// Inputs Validation
	if from == "" || pass == "" || to == "" || subject == "" || body == "" {
		template.ExecuteTemplate(w, "index.gohtml", TemplateData{ValidationError: true, Message: "Please fill all fields", Success: false, Method: "POST"})
		return
	}

	email_err := SendGmail_builtin(from, pass, to, subject, body)
	if email_err != nil {
		template.ExecuteTemplate(w, "index.gohtml", TemplateData{ValidationError: false, Message: fmt.Sprintf("Email not sent ):\nError:\n%v", err.Error()), Success: false, Method: "POST"})
	} else {
		template.ExecuteTemplate(w, "index.gohtml", TemplateData{ValidationError: false, Message: "Email sent (:", Success: true, Method: "POST"})
	}
}

// Apperantly gomail doesn't work
func SendGmail_gomail(from, pass, to, subject, body string) error {
	g := gomail.NewMessage()
	g.SetHeader("From", from)
	g.SetHeader("To", to)
	g.SetHeader("Subject", subject)
	g.SetBody("text/plain", body)
	server := gomail.NewDialer("smtp.gmail.com", 587, from, pass)

	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err := server.DialAndSend()
	return err
}

func SendGmail_builtin(from, pass, to, subject, body string) error {
	to_list := strings.Split(to, ";")
	message := []byte(body)

	auth := smtp.PlainAuth("", from, pass, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to_list, message)
	return err
}
