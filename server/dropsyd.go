package main
// as always a work in progres...

import(
	"fmt"
        "bufio"
	"strings"
	"net/http"
	"io/ioutil"
	"os"
	"path/filepath"
)

const STORAGE_ROOT = "storage"
const CERT_FILE = "config/server.crt"
const CERT_KEY_FILE = "config/server.key"
const PASSWD_FILE = "config/passwd"


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

func FileSave(req *http.Request, user string, dir string, file string) {
	var buf = req.Body;
	if dir == "" || file == "" { return }
	var path = filepath.Join(STORAGE_ROOT, user, dir)
        os.MkdirAll(path, 0777)

	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return
	}

	path = filepath.Join(STORAGE_ROOT, user, dir, file)

	fmt.Printf("create %s\n", path)
	f, err := os.Create(path)
	f.Write(bytes)
	f.Close()
}

func FileDelete(user string, dir string, file string) {
	if dir == "" || file == "" { return }
	var path = filepath.Join(STORAGE_ROOT, user, dir, file)
	fmt.Printf("remove %s\n", path)

	fi, err := os.Stat(path)
	if err != nil {
		return
	}

	mode := fi.Mode()
	if !mode.IsDir() {
		os.Remove(path)
		path := filepath.Join(STORAGE_ROOT, user, dir)
		for DirIsEmpty(path) {
			fmt.Printf("rmdir %s\n", path)
			os.Remove(path)
			end := strings.LastIndex(path, "/")
			if end < 0 { break }
			path = path[0:end]
		}
	}
}

func SendClientStatus(res http.ResponseWriter, value int) {
	res.WriteHeader(http.StatusOK)
	var format = fmt.Sprintf("status: %d\r\n\r\n", value)
	res.Write([]byte(format))
}

func AuthCheck(res http.ResponseWriter, username string, password string) (bool) {
	f, err := os.Open(PASSWD_FILE)
        if err != nil {
                fmt.Printf("FATAL: no credentials file found (%s)!\n", PASSWD_FILE)
                os.Exit(0)
        }

        fmt.Println(username, password)
        defer f.Close()

        r := bufio.NewReader(f) 

        for {
                bytes, err := r.ReadBytes('\n') 
                if err != nil {
                        return false
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

                if tmp_pass == password {
                        SendClientStatus(res, 0)
                        return true
                } else {
                        break
                }
        }

        SendClientStatus(res, 1)

	return false 
}

var header_list = []string {
        "username", "password", "action", "directory", "filename",
}

func ServerRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, "unsupported method!", http.StatusBadRequest)
		return
	}

        headers := make(map[string]string)

        for _, name := range(header_list) {
                headers[name] = req.Header.Get(name)
        }

	success := AuthCheck(res, headers["username"],headers["password"])
        if !success {
                return
        }

	switch headers["action"] {
        case "AUTH":
                // Already checked - do nothing.
                return
	case "ADD":
		FileSave(req, headers["username"], headers["directory"], headers["filename"])
	case "DEL":
		FileDelete(headers["username"], headers["directory"], headers["filename"])
	default:
		http.Error(res, "unknown req", http.StatusBadRequest)
	}
}

func Init() {
	if _, err := os.Stat(PASSWD_FILE); err != nil {
                // if os.IsNotExist(err) 
                fmt.Printf("FATAL: cannot read (%s)!\n", PASSWD_FILE)
                os.Exit(0)
        }
}

func Server() {
        Init()
	http.HandleFunc("/any", ServerRequest)
	if err := http.ListenAndServeTLS(":12345", CERT_FILE, CERT_KEY_FILE, nil); err != nil {
		fmt.Printf("FATAL: missing public/private key files!\n")
		os.Exit(0)
	}
}

func main() {
	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n")
	fmt.Printf("See: http://haxlab.org\n")
	fmt.Printf("Running: dropsyd daemon\n")

	Server()
}
