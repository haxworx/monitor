package main
import (
        "fmt"
        "net/http"
)

func SendClientStatus(res http.ResponseWriter, value int) (int) {
        var status = fmt.Sprintf("status: %d\r\n\r\n", value)
        res.Write([]byte(status))
	return value
}

