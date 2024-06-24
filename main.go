package main

import (
	"database/sql"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"rssagg/internal/database"
	"strconv"
	"time"
)

type apiConfig struct {
	DB            *database.Queries
	scrapeThreads int32
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln(err)
	}
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatalln("PORT environment variable not set")
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatalln("DB_URL environment variable not set")
	}

	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalln("can't connect to database", err)
	}

	scrapeThreads := int32(5)
	scrapeThreadsStr := os.Getenv("SCRAPE_THREADS")
	if scrapeThreadsStr != "" {
		threads, err := strconv.ParseInt(scrapeThreadsStr, 10, 32)
		if err != nil {
			log.Fatalln("bad value for SCRAPE_THREADS env var", err)
		}
		scrapeThreads = int32(threads)
	}

	apiCfg := apiConfig{
		DB:            database.New(conn),
		scrapeThreads: scrapeThreads,
	}

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	v1router := chi.NewRouter()
	v1router.Get("/healthz", handlerReadiness)
	v1router.Get("/err", handlerErr)
	v1router.Post("/users", apiCfg.handlerCreateUser)
	v1router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUserByApiKey))
	v1router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1router.Get("/feeds", apiCfg.handlerGetFeeds)
	v1router.Post("/feed-follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1router.Get("/feed-follows/{FeedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollow))
	v1router.Get("/feed-follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1router.Delete("/feed-follows/{FeedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))
	v1router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetUserPosts))
	router.Mount("/v1", v1router)
	server := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}
	log.Printf("Starting scraper\n")
	go weBeScrapin(apiCfg.DB, apiCfg.scrapeThreads, 1*time.Minute)
	log.Printf("Server starting on port %s\n", portString)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
