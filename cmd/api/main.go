package main

import (
	"github.com/denisakp/ogoune/internal/platform/bootstrap"
)

// @title Ogoune Public API
// @version 1.0
// @description Ogoune uptime monitoring — Public REST API v1.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	app := &bootstrap.App{}

	bootstrap.InitConfig(app)
	bootstrap.InitDatabase(app)
	bootstrap.InitMetrics(app)
	bootstrap.InitScheduler(app)
	bootstrap.InitWorker(app)
	bootstrap.InitServices(app)
	bootstrap.InitRouter(app)
	bootstrap.RunServer(app)
}
