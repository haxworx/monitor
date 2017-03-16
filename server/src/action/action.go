package action

import(
	"fmt"
	"strings"
	"io/ioutil"
	"io"
	"os"
	"path/filepath"
)

const STORAGE_ROOT = "storage"

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
	action string
	reader io.Reader
}

func New(reader io.Reader, action string) (*Action) {
	this := new(Action)
	this.action = action
	this.reader = reader
	
	return this
}

func (self *Action) Save(user string, dir string, file string) (bool) {

	var buf = self.reader;

	if user == "" || dir == "" || file == "" {
                return false
        }

	var path = filepath.Join(STORAGE_ROOT, user, dir)
        os.MkdirAll(path, 0777)

	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return false
	}

	path = filepath.Join(STORAGE_ROOT, user, dir, file)

	fmt.Printf("create %s\n", path)
	f, err := os.Create(path)
        if err != nil {
                fmt.Printf("ERROR: could not create: %s!\n", path)
		return false
        }

	f.Write(bytes)
	f.Close()

        return true
}

func (self *Action) Delete(user string, dir string, file string) (bool) {
	if user == "" || dir == "" || file == "" { return false }
	var path = filepath.Join(STORAGE_ROOT, user, dir, file)
	fmt.Printf("remove %s\n", path)

	fi, err := os.Stat(path)
	if err != nil {
		return false
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

	return true
}

func (self *Action) Process(user string, dir string, file string) (bool) {
	result := true
	switch self.action {
	case "ADD":
		result = self.Save(user, dir, file)
	case "DEL":
		result = self.Delete(user, dir, file)
	}

	return result
}



