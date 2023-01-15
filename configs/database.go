package configs

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"nerajima.com/NeraJima/models"
)

var (
	Database *gorm.DB
)

func InitDatabase() {
	db, err := gorm.Open(mysql.Open(EnvMySqlDNS()), &gorm.Config{})

	if err != nil {
		log.Fatal("Error connecting to database...")
		panic(err)
	}

	log.Println("Database connection established...")
	log.Println("Running migrations...")

	if err = setupJoinTables(db); err != nil {
		log.Fatal("Error during join table setup...")
		panic(err)
	}

	if err = db.AutoMigrate(
		&models.TemporaryObject{},
		&models.User{},
		&models.Profile{},
		&models.SearchHistory{},
		&models.Post{},
		&models.PostMedia{},
		&models.Comment{},
	); err != nil {
		log.Fatal("Error during migration...")
		panic(err)
	}

	log.Println("Migrations ran successfully!")

	Database = db
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
