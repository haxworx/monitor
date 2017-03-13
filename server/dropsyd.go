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

// 1 is good and 0 is bad!
func SendClientStatus(res http.ResponseWriter, value int) (int) {
        var status = fmt.Sprintf("status: %d\r\n\r\n", value)
        res.Write([]byte(status))
	return value
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
                return SendClientStatus(res, 0)
        }

        if self.Users[user_guess].password != pass_guess {
                return SendClientStatus(res, 0)
        }

        /* Works! */
        return SendClientStatus(res, 1)
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

type Action struct {
	res http.ResponseWriter
	req *http.Request
	action string
}

func (self *Action) Process(req *http.Request, res http.ResponseWriter, action string, user string, dir string, file string) {
	act := Action { req: req, res: res, action: action }
	switch act.action {
	case "ADD":
		act.Save(user, dir, file)
	case "DEL":
		act.Delete(user, dir, file)
	}
}

func (self *Action) Save(user string, dir string, file string) int {
	res := self.res
	req := self.req

	var buf = req.Body;

	if dir == "" || file == "" {
                return SendClientStatus(res, 0)
        }

	var path = filepath.Join(STORAGE_ROOT, user, dir)
        os.MkdirAll(path, 0777)

	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
                return SendClientStatus(res, 0)
	}

	path = filepath.Join(STORAGE_ROOT, user, dir, file)

	fmt.Printf("create %s\n", path)
	f, err := os.Create(path)
        if err != nil {
                fmt.Printf("FATAL: could not create: %s!\n", path)
                SendClientStatus(res, 0)
                os.Exit(1)
        }

	f.Write(bytes)
	f.Close()

        return SendClientStatus(res, 1)
}

func (self *Action) Delete(user string, dir string, file string) (int) {
	if dir == "" || file == "" { return 0 }
	var path = filepath.Join(STORAGE_ROOT, user, dir, file)
	fmt.Printf("remove %s\n", path)

	res := self.res
	fi, err := os.Stat(path)
	if err != nil {
                return SendClientStatus(res, 0)
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

        return SendClientStatus(res, 1)
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

	action := Action{}
	action.Process(req, res, headers["action"], headers["username"], headers["directory"], headers["filename"])
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

