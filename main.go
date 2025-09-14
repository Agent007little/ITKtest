package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ITKtest/config"
	"ITKtest/database"
	"ITKtest/internal/controller"
	"ITKtest/internal/repository"
	"ITKtest/internal/service"
	"ITKtest/responder"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db, false); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Initialize dependencies
	walletRepo := repository.NewWalletRepository(db)
	walletService := service.NewWalletService(walletRepo)
	resp := responder.NewJSONResponder()
	walletController := controller.NewWalletController(walletService, resp)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/wallet", walletController.HandleWalletOperation)
		r.Get("/wallets/{walletId}", walletController.GetWalletBalance)
	})

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
