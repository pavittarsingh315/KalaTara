package configs

import (
	"context"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"nerajima.com/NeraJima/models"
)

var (
	Database *gorm.DB
)

const (
	queryTimeout = time.Second
)

func InitDatabase() {
	db, err := gorm.Open(postgres.Open(EnvPostgresDNS()), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if sqlDb, err := db.DB(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	} else {
		sqlDb.SetMaxOpenConns(EnvDbMaxOpenConns())
		sqlDb.SetMaxIdleConns(EnvDbMaxIdleConns())
		sqlDb.SetConnMaxLifetime(time.Duration(EnvDbConnMaxLifetime()))
	}

	if !EnvProdActive() {
		migrate(db)
	}

	Database = db
}

func migrate(db *gorm.DB) {
	log.Println("Database connection established...")
	log.Println("Running migrations...")

	if err := setupJoinTables(db); err != nil {
		log.Fatalf("Error during join table setup: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.SearchHistory{},
		&models.Post{},
		&models.PostMedia{},
		&models.Comment{},
		&models.Notification{},
	); err != nil {
		log.Fatalf("Error during migration: %v", err)
	}

	log.Println("Migrations ran successfully!")
}

func setupJoinTables(db *gorm.DB) error {
	if err := db.SetupJoinTable(&models.Profile{}, "Followers", &models.ProfileFollower{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&models.Profile{}, "Subscribers", &models.ProfileSubscriber{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&models.Post{}, "Likes", &models.PostLike{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&models.Post{}, "Dislikes", &models.PostDislike{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&models.Post{}, "Bookmarks", &models.PostBookmark{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&models.Comment{}, "Likes", &models.CommentLike{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&models.Comment{}, "Dislikes", &models.CommentDislike{}); err != nil {
		return err
	}

	return nil
}

// Returns a context with a timeout of 1 second
func NewQueryContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), queryTimeout)
}
