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

// 1 is good and 0 is bad!
func SendClientActionStatus(res http.ResponseWriter, value int) (int) {
        var status = fmt.Sprintf("status: %d\r\n\r\n", value)
        res.Write([]byte(status))
	return value
}

func FileSave(req *http.Request, res http.ResponseWriter, user string, dir string, file string) int {
	var buf = req.Body;

	if dir == "" || file == "" { 
                return SendClientActionStatus(res, 0)
        }

	var path = filepath.Join(STORAGE_ROOT, user, dir)
        os.MkdirAll(path, 0777)

	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
                return SendClientActionStatus(res, 0)
	}

	path = filepath.Join(STORAGE_ROOT, user, dir, file)

	fmt.Printf("create %s\n", path)
	f, err := os.Create(path)
        if err != nil {
                fmt.Printf("FATAL: could not create: %s!\n", path)
                SendClientActionStatus(res, 0)
                os.Exit(1)
        }

	f.Write(bytes)
	f.Close()

        return SendClientActionStatus(res, 1)
}

func FileDelete(res http.ResponseWriter, user string, dir string, file string) (int) {
	if dir == "" || file == "" { return 0 }
	var path = filepath.Join(STORAGE_ROOT, user, dir, file)
	fmt.Printf("remove %s\n", path)

	fi, err := os.Stat(path)
	if err != nil {
                return SendClientActionStatus(res, 0)
	}

	mode := fi.Mode()
	if !mode.IsDir() {
		os.Remove(path)
		//path := filepath.Join(STORAGE_ROOT, user, dir)
		path := STORAGE_ROOT + "/" +  user + "/" + dir;
		for DirIsEmpty(path) {
			fmt.Printf("rmdir %s\n", path)
			os.Remove(path)
			end := strings.LastIndex(path, "/")
			if end < 0 { break }
			path = path[0:end]
		}
	}

        return SendClientActionStatus(res, 1)
}

type User struct {
        pid     int32
        username string
        password string
        name string
}

type Auth struct {
        Users map[string]User
        was_initialized bool 
}

var auth Auth

func (self *Auth) LoadFromFile() (bool) {
        f, err := os.Open(PASSWD_FILE)
        if err != nil {
                fmt.Printf("FATAL: no credentials file found (%s)!\n", PASSWD_FILE)
                os.Exit(0)
        }

        defer f.Close()

        r := bufio.NewReader(f)

        for {
                bytes, err := r.ReadBytes('\n')
                if err != nil { break }

                if bytes[0] == '#' { continue }

                line := string(bytes)

                eou := strings.Index(line, ":")
                tmp_user := line[0:eou]

                eop := strings.Index(line, "\n")
                tmp_pass := line[eou + 1:eop];
                var tmp User = User{}
                tmp.username = tmp_user
                tmp.password = tmp_pass
                self.Users[tmp_user] = tmp 
        }

        self.was_initialized = true

        return true
}

func (self *Auth) Check(res http.ResponseWriter, user_guess string, pass_guess string) (int) {
        if !self.was_initialized {
                self.Users = make(map[string]User)
                self.LoadFromFile()
        }

        if self.Users[user_guess].username != user_guess {
                return SendClientActionStatus(res, 0)
        }

        if self.Users[user_guess].password != pass_guess {
                return SendClientActionStatus(res, 0)
        }

        /* Works! */
        return SendClientActionStatus(res, 1)
}


func ServerRequest(res http.ResponseWriter, req *http.Request) {
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

	switch headers["action"] {
	case "ADD":
		FileSave(req, res, headers["username"], headers["directory"], headers["filename"])
	case "DEL":
		FileDelete(res, headers["username"], headers["directory"], headers["filename"])
	default:
	}
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

        auth = Auth{ was_initialized: false }
}

func Server() {
        Init()

	fmt.Printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n")
	fmt.Printf("See: http://haxlab.org\n")
	fmt.Printf("Running: dropsyd daemon\n")

	http.HandleFunc("/any", ServerRequest)
	if err := http.ListenAndServeTLS(":12345", CERT_FILE, CERT_KEY_FILE, nil); err != nil {
		fmt.Printf("FATAL: missing public/private key files!\n")
		os.Exit(0)
	}
}

func main() {
	Server()
}

