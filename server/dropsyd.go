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
		http.Error(response, "Not today", http.StatusBadRequest)
	};
}

func main() {
	http.HandleFunc("/any", PostRequest);
	fmt.Printf("(c) Copyright 2016. Al Poole <nestar@gmail.com>.\n");
	fmt.Printf("See: http://haxlab.org\n");
	fmt.Printf("Running: dropsyd daemon\n");
	http.ListenAndServe(":12345", nil);
}
