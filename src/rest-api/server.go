package restapi

import (
	"net/http"

	"github.com/cloud-barista/cb-ladybug/src/core/common"
	"github.com/cloud-barista/cb-ladybug/src/rest-api/router"
	"github.com/cloud-barista/cb-ladybug/src/utils/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/cloud-barista/cb-ladybug/src/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Server() {
	// Echo instance
	e := echo.New()

	// Echo middleware func
	e.Use(middleware.Logger())                             // Setting logger
	e.Use(middleware.Recover())                            // Recover from panics anywhere in the chain
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{ // CORS Middleware
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET(config.Config.BasePath+"/healthy", router.Healthy)

	g := e.Group(config.Config.BasePath+"/ns", common.NsValidate())

	// Routes
	g.GET("/:namespace/clusters", router.ListCluster)
	g.POST("/:namespace/clusters", router.CreateCluster)
	g.GET("/:namespace/clusters/:cluster", router.GetCluster)
	g.DELETE("/:namespace/clusters/:cluster", router.DestroyCluster)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
