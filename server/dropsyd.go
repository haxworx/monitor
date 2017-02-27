package main;

// as always a work in progres...

import(
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"os"
);

func ProcessPost(request *http.Request, dir string, file string) {
	var buf = request.Body;

        os.MkdirAll(dir, 0777);

	bytes, err := ioutil.ReadAll(buf);
	if err != nil {
		return;
	}

	var path = dir + "/" + file;

	fmt.Printf("create %s\n", path);
	f, err := os.Create(path);
	f.Write(bytes);
	f.Close();
}


func DirIsEmpty(directory string) bool {
	count := 0
	files, err := ioutil.ReadDir(directory);
	if err != nil {
		return false
	}

	count = len(files);
	if count > 0 {
		return false;
	}
	return true;
}

func AuthResponse(response http.ResponseWriter, value int) {
	response.WriteHeader(http.StatusOK);
	var format = fmt.Sprintf("status: %d\r\n\r\n", value);
	response.Write([]byte(format));
}

func CredentialsSet(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(response, "not supported!", http.StatusBadRequest);
	}

	request.ParseForm();

	username := request.FormValue("username")
	password := request.FormValue("password")

	if username == "" || password == "" {
		response.WriteHeader(http.StatusBadRequest);
		response.Write([]byte("<p>Empty fields!</p>"))
		return;
	}

	var path = "config/config.txt"
	output := fmt.Sprintf("username:%s\npassword:%s\n", username, password);
	f, _ := os.Create(path)
	f.Write([]byte(output))
	f.Close();

	filename := "html/success.html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest);
		response.Write([]byte("<p>missing internal files!</p>"))
		return;
	}

	response.WriteHeader(http.StatusOK);
	response.Write(body);
}

func CredentialsGet() (string, string) {
	var path = "config/config.txt"

	f, err := os.Open(path)
	if err != nil {
		os.Exit(1)
	}

	bytes, _ := ioutil.ReadAll(f)
	f.Close()

	data := string(bytes)

	usertag := "username:"

	idx := strings.Index(data, usertag) 
	start_user := data[idx + len(usertag):len(data)];
	end_user  := strings.Index(start_user, "\n");

	passtag := "password:"

	idx = strings.Index(data, passtag)
	start_pass := data[idx + len(passtag):len(data)]
	end_pass := strings.Index(start_pass, "\n");

	username := start_user[0:end_user];
	password := start_pass[0:end_pass];


	return username, password
}

func GenericRequest(response http.ResponseWriter, request *http.Request) {
	filename := "html/index.html"

	body, err := ioutil.ReadFile(filename);
	if err != nil {
		response.WriteHeader(http.StatusBadRequest);
	}

	response.WriteHeader(http.StatusOK);
	response.Write(body);
}

func PostRequest(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(response, "unsupported method!", http.StatusBadRequest);
		return;
	}

	username, password := CredentialsGet()

	var user = request.Header.Get("username");
	var pass = request.Header.Get("password");
	var file = request.Header.Get("filename");
	var action = request.Header.Get("action");
	var directory = request.Header.Get("directory");

	if user != username || pass != password {
		AuthResponse(response, 1);
		return;
	}

	switch action {
	case "ADD":
		if directory == "" || file == "" {
			return;
		}
		ProcessPost(request, directory, file)
	case "DEL":
		var path = directory + "/" + file
		fmt.Printf("remove %s\n", path)

		fi, err := os.Stat(path)
		if err != nil {
			return
		}

		mode := fi.Mode()
		if !mode.IsDir() {
			os.Remove(path)
			for DirIsEmpty(directory) {
				fmt.Printf("rmdir %s\n", directory);
				os.Remove(directory)
				end := strings.LastIndex(directory, "/");
				if end < 0 { break }
				directory = directory[0:end];
			}
		}
	case "AUTH":
		AuthResponse(response, 0)
	default:
		http.Error(response, "unknown request", http.StatusBadRequest)
	};
}

func main() {
	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n");
	fmt.Printf("See: http://haxlab.org\n");
	fmt.Printf("Running: dropsyd daemon\n");

	go func() {
		http.HandleFunc("/any", PostRequest);
		err := http.ListenAndServeTLS(":12345", "config/server.crt", "config/server.key", nil);
		if err != nil {
			fmt.Printf("missing public/private keys???\n");
			os.Exit(1);
		}
	}()
 	
	http.HandleFunc("/", GenericRequest)
	http.HandleFunc("/config", CredentialsSet);
	http.ListenAndServe(":80", nil);
}
