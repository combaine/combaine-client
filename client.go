package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/labstack/gommon/random"
)

var port int
var config string

func init() {
	flag.StringVar(&config, "config", "/etc/combaine/client-config.yaml", "Client configuration")
}

func main() {
	flag.Parse()

	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)

	cl, err := newConfigLoader(config, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "/ping ok")
	})
	e.GET("/exec", func(c echo.Context) error {
		return c.String(http.StatusOK, cl.tasksList())
	})

	e.GET("/exec/:task", func(c echo.Context) error {
		taskName := c.Param("task")
		rid := setRequestID(c)
		task, exists := cl.lookupTask(taskName)
		if !exists {
			return c.String(http.StatusNotFound, "Task "+taskName+" Not Found")

		}
		output, err := task.getOutput(rid, c)
		if err != nil {
			return c.String(http.StatusInternalServerError, "")
		}
		return c.String(http.StatusOK, output)
	})
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(cl.settings.Port)))
}

func setRequestID(c echo.Context) string {
	req := c.Request()
	res := c.Response()
	rid := req.Header.Get(echo.HeaderXRequestID)
	if rid == "" {
		rid = random.String(32)
	}
	res.Header().Set(echo.HeaderXRequestID, rid)
	return rid
}
