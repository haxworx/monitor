package main;

// as always a work in progres...

import(
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"os"
	"sync"
)

func SaveFile(request *http.Request, user string, dir string, file string) {
	var buf = request.Body;
	if dir == "" || file == "" { return }
	var path = user + "/" + dir
        os.MkdirAll(path, 0777)

	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return;
	}

	path = user + "/" + dir + "/" + file;

	fmt.Printf("create %s\n", path)
	f, err := os.Create(path)
	f.Write(bytes)
	f.Close()
}

func RemoveFile(user string, dir string, file string) {
	var path = user + "/" + dir + "/" + file
	if dir == "" || file == "" { return }
	fmt.Printf("remove %s\n", path)

	fi, err := os.Stat(path)
	if err != nil {
		return
	}

	mode := fi.Mode()
	if !mode.IsDir() {
		os.Remove(path)
		for DirIsEmpty(dir) {
			fmt.Printf("rmdir %s\n", dir)
			os.Remove(dir)
			end := strings.LastIndex(dir, "/")
			if end < 0 { break }
			dir = dir[0:end];
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

func CheckAuth(response http.ResponseWriter, value int) {
	response.WriteHeader(http.StatusOK)
	var format = fmt.Sprintf("status: %d\r\n\r\n", value)
	response.Write([]byte(format))
}

func CredentialsSet(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(response, "not supported!", http.StatusBadRequest)
	}

	request.ParseForm()

	username := request.FormValue("username")
	password := request.FormValue("password")

	if username == "" || password == "" {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("<p>Empty fields!</p>"))
		return;
	}

	var path = "config/config.txt"
	output := fmt.Sprintf("username:%s\npassword:%s\n", username, password)

	var mutex = &sync.Mutex{}

	mutex.Lock()
	f, _ := os.Create(path)
	f.Write([]byte(output))
	f.Close()
	mutex.Unlock()

	filename := "html/success.html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("<p>missing internal files!</p>"))
		return;
	}

	response.WriteHeader(http.StatusOK)
	response.Write(body)
}

func CredentialsGet() (string, string) {
	var path = "config/config.txt"

	f, _ := os.Open(path)
	bytes, _ := ioutil.ReadAll(f)
	f.Close()

	data := string(bytes)

	usertag := "username:"

	idx := strings.Index(data, usertag) 
	start_user := data[idx + len(usertag):len(data)];
	end_user  := strings.Index(start_user, "\n")

	passtag := "password:"

	idx = strings.Index(data, passtag)
	start_pass := data[idx + len(passtag):len(data)]
	end_pass := strings.Index(start_pass, "\n")

	username := start_user[0:end_user];
	password := start_pass[0:end_pass];


	return username, password
}

func GenericRequest(response http.ResponseWriter, request *http.Request) {
	filename := "html/index.html"

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
	}

	response.WriteHeader(http.StatusOK)
	response.Write(body)
}

func PostRequest(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(response, "unsupported method!", http.StatusBadRequest)
		return;
	}


	var user = request.Header.Get("username")
	var pass = request.Header.Get("password")
	var file = request.Header.Get("filename")
	var action = request.Header.Get("action")
	var directory = request.Header.Get("directory")

	username, password := CredentialsGet();

	if user != username || pass != password {
		CheckAuth(response, 1)
		return;
	}

	switch action {
	case "ADD":
		SaveFile(request, username, directory, file)
	case "DEL":
		RemoveFile(username, directory, file);
	case "AUTH":
		CheckAuth(response, 0)
	default:
		http.Error(response, "unknown request", http.StatusBadRequest)
	}
}

func SettingsServer() {
	fmt.Printf("Configure this instance at: http://localhost\n")
	http.HandleFunc("/", GenericRequest)
	http.HandleFunc("/config", CredentialsSet)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Printf("unable to bind to port 80\n")
		os.Exit(0)
	}
}

func MainServer() {
	http.HandleFunc("/any", PostRequest)
	err := http.ListenAndServeTLS(":12345", "config/server.crt", "config/server.key", nil)
	if err != nil {
		fmt.Printf("missing public/private keys???\n")
		os.Exit(0)
	}
}

func main() {
	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n")
	fmt.Printf("See: http://haxlab.org\n")
	fmt.Printf("Running: dropsyd daemon\n")

	go func() {
		SettingsServer();
	}()

	MainServer();
}
