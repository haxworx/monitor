package main
// as always a work in progres...

import(
	"fmt"
        "bufio"
	"strings"
	"net/http"
	"io/ioutil"
	"os"
	"sync"
)

const PASSWD_FILE = "config/passwd"

func SaveToDisk(req *http.Request, user string, dir string, file string) {
	var buf = req.Body;
	if dir == "" || file == "" { return }
	var path = user + "/" + dir
        os.MkdirAll(path, 0777)

	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return
	}

	path = user + "/" + dir + "/" + file

	fmt.Printf("create %s\n", path)
	f, err := os.Create(path)
	f.Write(bytes)
	f.Close()
}

func DelFromDisk(user string, dir string, file string) {
	if dir == "" || file == "" { return }
	var path = user + "/" + dir + "/" + file
	fmt.Printf("remove %s\n", path)

	fi, err := os.Stat(path)
	if err != nil {
		return
	}

	mode := fi.Mode()
	if !mode.IsDir() {
		os.Remove(path)
		path := user + "/" + dir
		for DirIsEmpty(path) {
			fmt.Printf("rmdir %s\n", path)
			os.Remove(path)
			end := strings.LastIndex(path, "/")
			if end < 0 { break }
			path = path[0:end]
		}
	}
}

func DirIsEmpty(directory string) bool {
	count := 0
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return false
	}

	count = len(files)
	if count > 0 {
		return false;
	}
	return true;
}

func SendClientStatus(res http.ResponseWriter, value int) {
	res.WriteHeader(http.StatusOK)
	var format = fmt.Sprintf("status: %d\r\n\r\n", value)
	res.Write([]byte(format))
}

func SettingsUpdate(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, "not supported!", http.StatusBadRequest)
	}

	req.ParseForm()

	username := req.FormValue("username")
	password := req.FormValue("password")

	if username == "" || password == "" {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("<p>Empty fields!</p>"))
		return
	}

	var path = PASSWD_FILE
	output := fmt.Sprintf("%s:%s\n", username, password)

	var mutex = &sync.Mutex{}

	mutex.Lock()
	f, _ := os.Create(path)
	f.Write([]byte(output))
	f.Close()
	mutex.Unlock()

	filename := "html/success.html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("<p>missing internal files!</p>"))
		return;
	}

	res.WriteHeader(http.StatusOK)
	res.Write(body)
}

type Credentials_t struct {
        username string;
        password string;
}

func CheckAuth(res http.ResponseWriter, username string, password string) (Credentials_t, error) {
	var path = PASSWD_FILE

        var entry Credentials_t

	f, err := os.Open(path)
        if err != nil {
                return entry, err
        }

        defer f.Close()

        r := bufio.NewReader(f) 

        for {
                bytes, err := r.ReadBytes('\n') 
                if err != nil {
                        return entry, err
                }

                if bytes[0] == '#' { continue }

                line := string(bytes)

                eou := strings.Index(line, ":")
                tmp_user := line[0:eou]
                if tmp_user != username {
                        continue
                }

                eop := strings.Index(line, "\n")

                tmp_pass := line[eou + 1:eop];

                entry.username = tmp_user
	        entry.password = tmp_pass
                if tmp_pass != password {
                        SendClientStatus(res, 1)
                        return entry, nil
                } else {
                        break
                }
        }

        SendClientStatus(res, 0)

	return entry, nil
}

func SettingsIndexShow(res http.ResponseWriter, req *http.Request) {
	filename := "html/index.html"

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
	}

	res.WriteHeader(http.StatusOK)
	res.Write(body)
}

func MainServerRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, "unsupported method!", http.StatusBadRequest)
		return
	}

	var user_guess = req.Header.Get("username")
	var pass_guess = req.Header.Get("password")
	var file = req.Header.Get("filename")
	var action = req.Header.Get("action")
	var directory = req.Header.Get("directory")

	credits, err := CheckAuth(res, user_guess, pass_guess)
        if err != nil {
                return
        }

	switch action {
        case "AUTH":
                return
	case "ADD":
		SaveToDisk(req, credits.username, directory, file)
	case "DEL":
		DelFromDisk(credits.username, directory, file)
	default:
		http.Error(res, "unknown req", http.StatusBadRequest)
	}
}

func SettingsServer() {
	fmt.Printf("Configure this instance at: http://localhost:8080\n")
	http.HandleFunc("/", SettingsIndexShow)
	http.HandleFunc("/config", SettingsUpdate)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("unable to bind to port 8080\n")
		os.Exit(0)
	}
}

func MainServer() {
	http.HandleFunc("/any", MainServerRequest)
	if err := http.ListenAndServeTLS(":12345", "config/server.crt", "config/server.key", nil); err != nil {
		fmt.Printf("FATAL: missing public/private keys???\n")
		os.Exit(0)
	}
}

func main() {
	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n")
	fmt.Printf("See: http://haxlab.org\n")
	fmt.Printf("Running: dropsyd daemon\n")

	go SettingsServer()

	MainServer()
}
