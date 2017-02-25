package main;

import(
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
);

func ProcessPost(request *http.Request, dir string, file string) {
	var buf = request.Body;

        os.MkdirAll(dir, 0777);
	fmt.Printf("making %s\n", dir);

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


func RemoveEmptyDirs(directory string) {
	// do this later potato!
}

func RespondCode(response http.ResponseWriter, value int) {
	response.WriteHeader(http.StatusOK);
	var format = fmt.Sprintf("status: %d\r\n\r\n", value);
	response.Write([]byte(format));
}

func PostRequest(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(response, "POST ONLY", http.StatusBadRequest);
		return;
	}

	var user = request.Header.Get("username");
	var pass = request.Header.Get("password");
	var file = request.Header.Get("filename");
	var action = request.Header.Get("action");
	var directory = request.Header.Get("directory");

	if user != "username" || pass != "password" {
		RespondCode(response, 1);
		return;
	} else {
	}

	switch action {
	case "ADD":
		if directory == "" || file == "" {
			return;
		}

		ProcessPost(request, directory, file);
	case "DEL":
		var path = directory + "/" + file;
		os.Remove(path);
		RemoveEmptyDirs(directory);
	case "AUTH":
		RespondCode(response, 0);
		return;
	default:
		http.Error(response, "BOGUS ACTION", http.StatusBadRequest);
	};
}

func main() {
	http.HandleFunc("/any", PostRequest);
	fmt.Printf("(c) Copyright 2016. Al Poole <nestar@gmail.com>.\n");
	fmt.Printf("See: http://haxlab.org\n");
	fmt.Printf("Running: dropsyd daemon\n");
	http.ListenAndServe(":12345", nil);
}
