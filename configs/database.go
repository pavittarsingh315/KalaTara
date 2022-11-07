package configs

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"nerajima.com/NeraJima/models"
)

var (
	Database *gorm.DB
)

func InitDatabase() {
	db, err := gorm.Open(sqlite.Open("dev.db"), &gorm.Config{})

	if err != nil {
		log.Fatal("Error connecting to database...")
		panic(err)
	}

	log.Println("Database connection established...")
	log.Println("Running migrations...")

	if err = db.AutoMigrate(&models.TemporaryObject{}, &models.User{}, &models.Profile{}); err != nil {
		log.Fatal("Error during migration...")
		panic(err)
	}

	log.Println("Migrations ran successfully!")

	Database = db
}
