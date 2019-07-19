package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

var port int
var config string

func init() {
	flag.IntVar(&port, "port", 8080, "listen port")
	flag.StringVar(&config, "tasks", "/etc/combaine/client-tasks.toml", "Client tasks configuration")
}

func main() {
	flag.Parse()

	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)

	cl, err := newConfigLoader(config, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Use(middleware.Logger())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "/ping ok")
	})

	e.GET("/exec/:task", func(c echo.Context) error {
		taskName := c.Param("task")
		task, exists := cl.lookup(taskName)
		if !exists {
			return c.String(http.StatusNotFound, "Task "+taskName+" Not Found")

		}
		c.Logger().Infof("Run task %#v", task)
		return c.String(http.StatusOK, task.getOutput(e.Logger))
	})
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(port)))
}
