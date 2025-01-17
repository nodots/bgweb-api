package main

import (
	"bgweb-api/internal/api"
	"bgweb-api/internal/gnubg"
	"bgweb-api/internal/openapi"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/flowchartsman/swaggerui"
)

func main() {
	var datadir = flag.String("datadir", "./cmd/bgweb-api/data", "Folder containing gnubg data")
	var port = flag.Int("port", 8080, "Port for HTTP server")
	flag.Parse()

	if err := gnubg.Init(os.DirFS(*datadir)); err != nil {
		panic(fmt.Errorf("failed to initialize gnubg: %w", err))
	}

	swagger, err := openapi.GetSwagger()
	if err != nil {
		panic(fmt.Errorf("failed to get swagger: %w", err))
	}

	e := echo.New()

	e.Use(echomiddleware.CORS())

	// Log all requests
	e.Use(echomiddleware.Logger())

	v1 := e.Group("/api/v1")
	{
		// Check all requests against the OpenAPI schema.
		v1.Use(middleware.OapiRequestValidator(swagger))

		// register routes which were generated by openapi code generator
		// first, need our instance of the generated ServerInterface
		var api BackgammonWebAPI
		// then, register the routes
		openapi.RegisterHandlers(v1, &api)
	}

	// serve swagger ui
	{
		data, _ := swagger.MarshalJSON()
		e.GET("/swagger/*any", echo.WrapHandler(http.StripPrefix("/swagger", swaggerui.Handler(data))))
	}

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/swagger/")
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", *port)))
}

type BackgammonWebAPI struct {
}

func (*BackgammonWebAPI) PostGetmoves(c echo.Context) (err error) {
	var args openapi.MoveArgs

	// unmarshal body
	if err = c.Bind(&args); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// process logic
	moves, err := api.GetMoves(args)

	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	return c.JSON(http.StatusOK, moves)
}
