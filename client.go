package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8080, "listen port")
}

func main() {
	flag.Parse()

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "/ping ok")
	})
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(port)))
}
