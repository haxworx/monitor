package main

import(
        "fmt"
        "net/http"
        "./action"
	"./auth"
)

const URL = "/any"
const CERT_FILE = "config/server.crt"
const CERT_KEY_FILE = "config/server.key"

var authSystem *auth.Auth

func HandleRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		var header_list = []string {
			"username", "password", "action", "directory", "filename",
		}

		headers := make(map[string]string)

		for _, name := range(header_list) {
			headers[name] = req.Header.Get(name)
		}

		if success := authSystem.Check(res, headers["username"],headers["password"]); success != true {
			return
		}

		action := action.New(req, res, headers["action"])
		action.Process(headers["username"], headers["directory"], headers["filename"])
	} else {
                http.Error(res, "unsupported method!", http.StatusBadRequest)
	}
}

func Server() {
        http.HandleFunc(URL, HandleRequest)
        if err := http.ListenAndServeTLS(":12345", CERT_FILE, CERT_KEY_FILE, nil); err != nil {
                fmt.Println("FATAL: missing public/private key files!\n")
        }
}

func ShowAbout() {
	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n")
        fmt.Printf("See: http://haxlab.org\n")
        fmt.Printf("Running: dropsyd daemon\n")
}

func main() {
        authSystem = auth.New()
	ShowAbout()
	Server()
}

