package main
// as always a work in progres...

import(
	"fmt"
        "bufio"
	"strings"
	"net/http"
	"os"
        "time"
)

type User struct {
        uid int32
        username string
        password string
        name string
}

type Auth struct {
        Users map[string]User
        last_update int64
        have_auth_data bool 
}

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

        self.have_auth_data = true
        self.last_update = time.Now().Unix()

        return true
}

func (self *Auth) Check(res http.ResponseWriter, user_guess string, pass_guess string) (int) {
        if !self.have_auth_data {
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

