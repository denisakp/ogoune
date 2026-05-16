package bootstrap

import (
	"log/slog"
	"os"

	ogounedocs "github.com/denisakp/ogoune/docs"
	"github.com/denisakp/ogoune/internal/config"
	icmppkg "github.com/denisakp/ogoune/internal/icmp"
	"github.com/denisakp/ogoune/pkg/crypto"
	"github.com/denisakp/ogoune/pkg/logger"
)

// InitConfig loads configuration, initializes the logger, and validates the crypto key.
func InitConfig(app *App) {
	cfg := config.MustInit()
	app.Cfg = &cfg

	// Set Swagger host dynamically
	swaggerHost := config.GetEnv("SWAGGER_HOST", "localhost:"+cfg.Port)
	ogounedocs.SwaggerInfo.Host = swaggerHost
	ogounedocs.SwaggerInfo.Version = AppVersion

	// Initialize structured logger
	l := logger.New(cfg.LogFormat, cfg.LogLevel)
	slog.SetDefault(l)

	slog.Info("starting Ogoune application")
	LogStartupEdition()

	// Fail fast if APP_SECRET_KEY is missing or malformed
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	if err := crypto.ValidateKey(); err != nil {
		slog.Error("crypto key validation failed", "error", err)
		os.Exit(1)
	}
	slog.Info("encryption key validated")

	// Detect ICMP capability at startup
	icmpCapability := icmppkg.Detect()
	logICMPCapabilityState(cfg.EnableICMP, icmpCapability)

	slog.Info("authentication configured", "email", cfg.AuthEmail)
}
