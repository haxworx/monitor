package main
// as always a work in progres...

import(
	"fmt"
        "bufio"
	"strings"
	"net/http"
	"io/ioutil"
	"os"
)

const STORAGE_ROOT = "storage"
const CERT_FILE = "config/server.crt"
const CERT_KEY_FILE = "config/server.key"
const PASSWD_FILE = "config/passwd"

type Credentials_t struct {
        username string;
        password string;
        // OTHER STUFF???
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

func FileSave(req *http.Request, user string, dir string, file string) {
	var buf = req.Body;
	if dir == "" || file == "" { return }
	var path = STORAGE_ROOT + "/" + user + "/" + dir
        os.MkdirAll(path, 0777)

	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return
	}

	path = STORAGE_ROOT + "/" + user + "/" + dir + "/" + file

	fmt.Printf("create %s\n", path)
	f, err := os.Create(path)
	f.Write(bytes)
	f.Close()
}

func FileDelete(user string, dir string, file string) {
	if dir == "" || file == "" { return }
	var path = STORAGE_ROOT + "/" + user + "/" + dir + "/" + file
	fmt.Printf("remove %s\n", path)

	fi, err := os.Stat(path)
	if err != nil {
		return
	}

	mode := fi.Mode()
	if !mode.IsDir() {
		os.Remove(path)
		path := STORAGE_ROOT + "/" + user + "/" + dir
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

func AuthCheck(res http.ResponseWriter, username string, password string) (Credentials_t, error) {
	var path = PASSWD_FILE

        var entry Credentials_t

	f, err := os.Open(path)
        if err != nil {
                fmt.Printf("FATAL: no credentials file found (%s)!\n", PASSWD_FILE)
                os.Exit(0)
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

	credits, err := AuthCheck(res, user_guess, pass_guess)
        if err != nil {
                return
        }

	switch action {
        case "AUTH":
                // Already checked - do nothing.
                return
	case "ADD":
		FileSave(req, credits.username, directory, file)
	case "DEL":
		FileDelete(credits.username, directory, file)
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

func MainServer() {
        Init()
	http.HandleFunc("/any", MainServerRequest)
	if err := http.ListenAndServeTLS(":12345", CERT_FILE, CERT_KEY_FILE, nil); err != nil {
		fmt.Printf("FATAL: missing public/private key files!\n")
		os.Exit(0)
	}
}

func main() {
	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n")
	fmt.Printf("See: http://haxlab.org\n")
	fmt.Printf("Running: dropsyd daemon\n")

	MainServer()
}
