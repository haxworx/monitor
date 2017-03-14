package main
// as always a work in progres...

import(
	"fmt"
	"net/http"
	"os"
)

const STORAGE_ROOT = "storage"
const CERT_FILE = "config/server.crt"
const CERT_KEY_FILE = "config/server.key"
const PASSWD_FILE = "config/passwd"

var auth Auth

func ClientRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, "unsupported method!", http.StatusBadRequest)
		return
	}

	var header_list = []string {
		"username", "password", "action", "directory", "filename",
	}

        headers := make(map[string]string)

        for _, name := range(header_list) {
                headers[name] = req.Header.Get(name)
        }

	success := auth.Check(res, headers["username"],headers["password"])
        if (success != 1) { return }

	action := Action{ req: req, res: res, action: headers["action"]}
	action.Process(headers["username"], headers["directory"], headers["filename"])
}

func Init() {
	if _, err := os.Stat(PASSWD_FILE); err != nil {
                if os.IsNotExist(err) {
			fmt.Printf("FATAL: (%s) does not exist!\n", PASSWD_FILE)
		} else {
			fmt.Printf("FATAL: cannot stat() (%s)!\n", PASSWD_FILE)
		}
                os.Exit(0)
        }

        auth = Auth{ have_auth_data: false }
}

func Server() {
        Init()

	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n")
	fmt.Printf("See: http://haxlab.org\n")
	fmt.Printf("Running: dropsyd daemon\n")

	http.HandleFunc("/any", ClientRequest)
	if err := http.ListenAndServeTLS(":12345", CERT_FILE, CERT_KEY_FILE, nil); err != nil {
		fmt.Printf("FATAL: missing public/private key files!\n")
		os.Exit(0)
	}
}

func main() {
	Server()
}

