package auth
// as always a work in progres...

import(
	"fmt"
        "bufio"
	"strings"
	"os"
        "time"
)

const PASSWD_FILE = "config/passwd"

type User struct {
        uid int32
        username string
        password string
        name string
}

type Auth struct {
        Users map[string]User
        last_update int64
        initialized bool 
}

func New() (*Auth) {
	this := new(Auth)
	this.initialized = false
        
	if _, err := os.Stat(PASSWD_FILE); err != nil {
        	fmt.Printf("%s!\n", err)
                os.Exit(0)
        }

	return this
}

func (self *Auth) LoadFromFile() (bool) {
        f, err := os.Open(PASSWD_FILE)
        if err != nil {
                fmt.Printf("%s!\n", err)
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

        self.initialized = true
        self.last_update = time.Now().Unix()

        return true
}

func (self *Auth) Check(user_guess string, pass_guess string) (bool) {
        if !self.initialized {
                self.Users = make(map[string]User)
                self.LoadFromFile()
        }

        if self.Users[user_guess].username != user_guess {
		return false
        }

        if self.Users[user_guess].password != pass_guess {
		return false
        }

        /* Works! */
	return true
}

