package lh_test

import (
	"fmt"
	"math"
	"net/http"

	"github.com/cloudinterfaces/lh"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

type Math struct{}

func (m *Math) Add(_ *http.Request, args *[]float64, sum *float64) error {
	for _, a := range *args {
		*sum += a
	}
	return nil
}

func (m *Math) Multiply(_ *http.Request, args *[]float64, product *float64) error {
	for i, a := range *args {
		if a == 0 {
			*product = 0
			return nil
		}
		if i == 0 {
			*product = a
			continue
		}
		*product *= a
	}
	return nil
}

func (m *Math) Divide(_ *http.Request, args *[]float64, div *float64) error {
	if len(*args) < 2 {
		return fmt.Errorf("At least 2 arguments required")
	}
	for _, f := range (*args)[1:] {
		if math.Abs(f) == 0 {
			return fmt.Errorf("Args includes a division by zero")
		}
	}
	*div = (*args)[0] / (*args)[1]
	for _, f := range (*args)[2:] {
		*div /= f
	}
	return nil
}

func Example() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, ok := w.(lh.Lambda)
		switch ok {
		case true:
			fmt.Fprintf(w, "Hello from the Lambda environment! via %s\n", r.Host)
		default:
			fmt.Fprintln(w, "Not running in Lambda environment")
		}
	})
	lh.ServeHTTP(nil)
}

func Example_complicated() {
	server := rpc.NewServer() // github.com/gorilla/rpc
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterService(&Math{}, "Math")
	http.Handle("/rpc", server)
	g := gin.Default()
	g.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	http.Handle("/", g)
	http.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "../error", 307)
	})
	http.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "This is an error", 500)
	})
	lh.ServeHTTP(nil)
}
