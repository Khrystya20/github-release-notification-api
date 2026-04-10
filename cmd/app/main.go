package main

import (
	"log"
	"os"
	"time"

	githubclient "github-release-notification-api/internal/client/github"
	mailclient "github-release-notification-api/internal/client/mail"
	"github-release-notification-api/internal/config"
	"github-release-notification-api/internal/db"
	"github-release-notification-api/internal/handler"
	postgresrepo "github-release-notification-api/internal/repository/postgres"
	"github-release-notification-api/internal/scheduler"
	"github-release-notification-api/internal/service"
)

func main() {
	cfg := config.Load()

	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	repoRepository := postgresrepo.NewRepositoryRepository(database)
	subscriptionRepository := postgresrepo.NewSubscriptionRepository(database)

	githubClient := githubclient.NewClient(cfg.GitHubToken)
	mailSender := mailclient.NewSender(cfg)

	subscriptionService := service.NewSubscriptionService(
		repoRepository,
		subscriptionRepository,
		githubClient,
		mailSender,
	)

	scannerService := service.NewScannerService(
		repoRepository,
		subscriptionRepository,
		githubClient,
		mailSender,
	)

	interval, err := time.ParseDuration(cfg.ScanInterval)
	if err != nil {
		log.Fatalf("invalid SCAN_INTERVAL: %v", err)
	}

	scheduler.Start(scannerService, interval)

	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)
	router := handler.SetupRouter(subscriptionHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.AppPort
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("server started on port %s", cfg.AppPort)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
