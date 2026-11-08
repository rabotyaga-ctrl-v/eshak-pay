package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"smartwallet/internal/config"
	"smartwallet/internal/database"
	"smartwallet/internal/handler"
	"smartwallet/internal/repository"
	"smartwallet/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	accountService := service.NewAccountService(pool)

	userService := service.NewUserService(
		userRepo,
		pool,
		accountService,
	)

	sessionRepo := repository.NewSessionRepository(pool)
	sessionService := service.NewSessionService(sessionRepo)
	walletService := service.NewWalletService(pool)

	userHandler := handler.NewUserHandler(userService, sessionService)
	accountHandler := handler.NewAccountHandler(accountService)
	transferService := service.NewTransferService(pool)
	transferHandler := handler.NewTransferHandler(transferService)
	walletHandler := handler.NewWalletHandler(walletService)

	// фронт
	http.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/register.html")
	})

	http.HandleFunc("POST /register", userHandler.Register)
	http.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/login.html")
	})

	http.HandleFunc("POST /login", userHandler.Login)
	http.Handle(
		"GET /me",
		handler.AuthMiddleware(
			sessionService,
			http.HandlerFunc(userHandler.Me),
		),
	)

	http.Handle(
		"GET /account",
		handler.AuthMiddleware(
			sessionService,
			http.HandlerFunc(accountHandler.GetPrimaryAccount),
		),
	)

	http.HandleFunc("/logout", userHandler.Logout)
	http.Handle(
		"/transfer",
		handler.AuthMiddleware(
			sessionService,
			http.HandlerFunc(transferHandler.Transfer),
		),
	)
	http.Handle(
		"/transactions",
		handler.AuthMiddleware(
			sessionService,
			http.HandlerFunc(walletHandler.GetHistory),
		),
	)
	http.HandleFunc("GET /dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/dashboard.html")
	})

	http.Handle(
		"/transfer-page",
		handler.AuthMiddleware(
			sessionService,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "./web/static/transfer.html")
			}),
		),
	)

	http.Handle(
		"/transactions-page",
		handler.AuthMiddleware(
			sessionService,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "./web/static/transactions.html")
			}),
		),
	)

	http.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/about.html")
	})

	var dbName string
	err = pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbName)
	if err != nil {
		log.Fatalf("query database name: %v", err)
	}

	fmt.Printf("Connected to PostgreSQL database: %s\n", dbName)

	log.Println("Server started on :8080")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
