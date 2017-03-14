package main
// as always a work in progres...

import(
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"os"
	"path/filepath"
)

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

func (self *Action) Process(user string, dir string, file string) {
	switch self.action {
	case "ADD":
		self.Save(user, dir, file)
	case "DEL":
		self.Delete(user, dir, file)
	}
}



