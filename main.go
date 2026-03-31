package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/spf13/cobra"

	"podux/internal/application/dashboard"
	"podux/internal/application/importer"
	"podux/internal/application/integration"
	"podux/internal/application/monitoring"
	"podux/internal/application/proxy"
	"podux/internal/application/server"
	"podux/internal/application/system"
	cftunnel "podux/internal/application/tunnel/cloudflare"
	frptunnel "podux/internal/application/tunnel/frp"
	"podux/internal/application/version"
	"podux/internal/hooks"
	"podux/internal/infrastructure/persistence"
	httphandler "podux/internal/interfaces/http"
	_ "podux/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

//go:embed all:pb_public
var pbPublicDir embed.FS

func main() {
	app := pocketbase.New()

	// enable auto creation of migration files when using the "migrate" command
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	// Infrastructure (Repository implementations)
	serverRepo := persistence.NewServerRepository(app)
	proxyRepo := persistence.NewProxyRepository(app)
	tunnelRepo := persistence.NewTunnelRepository(app)

	// Application Services
	serverService := server.NewService(app, serverRepo)
	proxyService := proxy.NewService(app, proxyRepo)
	frpcService := frptunnel.NewProvider(app, serverRepo, proxyRepo, tunnelRepo)
	dashboardService := dashboard.NewService(app, serverService, proxyService, tunnelRepo)
	settingsService := system.NewService(app)
	geoService := monitoring.NewGeoService(app)
	metricsService := monitoring.NewMetricsService(app)
	metricsScheduler := monitoring.NewMetricsScheduler(app, metricsService)
	networkMonitorService := monitoring.NewMonitorService(app, geoService, metricsService)
	importService := importer.NewService(app)
	versionService := version.NewService(app)
	//githubService := github.NewService(app)

	cloudflareService := integration.NewCloudflareService(app)
	cloudflareProvider := cftunnel.NewProvider(app, proxyRepo, tunnelRepo)

	// HTTP Handlers
	//githubHandler := httphandler.NewGithubHandler(app, githubService)
	integrationHandler := httphandler.NewIntegrationHandler(app)
	frpcHandler := httphandler.NewTunnelHandler(app, frpcService, "/api/frpc")
	cloudflareHandler := httphandler.NewTunnelHandler(app, cloudflareProvider, "/api/cloudflare")
	versionHandler := httphandler.NewVersionHandler(app, versionService)
	dashboardHandler := httphandler.NewDashboardHandler(dashboardService)
	systemHandler := httphandler.NewSystemHandler(app, settingsService)
	importHandler := httphandler.NewImportHandler(app, importService)
	serverHandler := httphandler.NewServerHandler(app, metricsService)

	// Register all record hooks
	hooks.Register(app, cloudflareService, frpcService)

	// register custom routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Reset all tunnel statuses to inactive on startup
		tunnelRepo.ResetAllStatus()
		// Reset all proxy statuses to offline on startup
		proxyService.ResetAllProxyBootStatus()

		// Auto-start tunnels (non-blocking)
		go frpcService.AutoStart()
		go cloudflareProvider.AutoStart()

		// Start network monitoring service using intervals from settings
		latencyInterval, geoInterval := settingsService.GetMonitoringIntervals()
		networkMonitorService.Start(latencyInterval, geoInterval)
		metricsScheduler.Register()

		// Register routes for each module
		//githubHandler.RegisterHandlers(e)
		integrationHandler.RegisterHandlers(e)
		frpcHandler.RegisterHandlers(e)
		cloudflareHandler.RegisterHandlers(e)
		versionHandler.RegisterHandlers(e)
		dashboardHandler.RegisterHandlers(e)
		systemHandler.RegisterHandlers(e)
		importHandler.RegisterHandlers(e)
		serverHandler.RegisterHandlers(e)

		// Serve static files from embedded FS
		content, err := fs.Sub(pbPublicDir, "pb_public")
		if err != nil {
			return err
		}

		// Check if index.html exists to ensure frontend is properly embedded
		if _, err := content.Open("index.html"); err != nil {
			log.Printf("WARNING: index.html not found in embedded assets! Frontend may not load. Error: %v", err)
		}

		e.Router.GET("/{path...}", apis.Static(content, true))

		return e.Next()
	})

	// Overwrite the default serve command to suppress the default startup banner
	// and use our custom logs
	for _, cmd := range app.RootCmd.Commands() {
		if cmd.Use == "serve" {
			cmd.Run = nil
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				domains := args

				httpAddr, _ := cmd.Flags().GetString("http")
				httpsAddr, _ := cmd.Flags().GetString("https")
				origins, _ := cmd.Flags().GetStringSlice("origins")

				// replicate default logic from pocketbase
				if httpAddr == "" {
					if len(domains) > 0 {
						httpAddr = "0.0.0.0:80"
					} else {
						httpAddr = "127.0.0.1:8090"
					}
				}

				if httpsAddr == "" {
					if len(domains) > 0 {
						httpsAddr = "0.0.0.0:443"
					}
				}

				log.Println("Starting PocketBase server...")
				log.Println("----------------------------------------------------------------")
				log.Println("Frontend: http://" + httpAddr + "/")
				log.Println("Admin UI: http://" + httpAddr + "/_/")
				log.Println("----------------------------------------------------------------")

				return apis.Serve(app, apis.ServeConfig{
					HttpAddr:           httpAddr,
					HttpsAddr:          httpsAddr,
					ShowStartBanner:    false,
					AllowedOrigins:     origins,
					CertificateDomains: domains,
				})
			}
			break
		}
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
