package main

import (
	"time"

	"github.com/denisakp/ogoune/internal/platform/bootstrap"
)

// @title Ogoune Public API
// @version 1.0
// @description Ogoune uptime monitoring — Public REST API v1.
// @host localhost:9596
// @schemes http https
// @BasePath /api/v1
// @externalDocs.description Ogoune repository
// @externalDocs.url https://github.com/denisakp/ogoune
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

	stopAggregator := bootstrap.StartUptimeAggregator(app, 5*time.Minute)
	defer stopAggregator()

	bootstrap.RunServer(app)
}
