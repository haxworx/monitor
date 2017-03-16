package main

import(
        "fmt"
        "net/http"
	"./auth"
        "./action"
)

const POST_PATH = "/any"
const CERT_FILE = "config/server.crt"
const CERT_KEY_FILE = "config/server.key"

var authSystem *auth.Auth

func ClientSendStatus(res http.ResponseWriter, ok bool) (int) {
	value := 0
	if ok {
		value = 1
	} 

        var status = fmt.Sprintf("status: %d\r\n\r\n", value)
        res.Write([]byte(status))
        return value
}

func HandleRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		var header_list = []string {
			"username", "password", "action", "directory", "filename",
		}

		headers := make(map[string]string)

		for _, name := range(header_list) {
			headers[name] = req.Header.Get(name)
		}

		status := authSystem.Check(headers["username"], headers["password"])
		ClientSendStatus(res, status)
		if status != true {
			return
		}

		action := action.New(req.Body, headers["action"])
		status = action.Process(headers["username"], headers["directory"], headers["filename"])
		ClientSendStatus(res, status)

	} else {
                http.Error(res, "unsupported method!", http.StatusBadRequest)
	}
}

func Server() {
        http.HandleFunc(POST_PATH, HandleRequest)
        if err := http.ListenAndServeTLS(":12345", CERT_FILE, CERT_KEY_FILE, nil); err != nil {
                fmt.Printf("%s!\n", err)
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
