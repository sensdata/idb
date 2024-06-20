package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sensdata/idb/core/utils"
)

var (
	db   *sql.DB
	once sync.Once
)

func Init(dataSourceName string) error {
	var err error
	once.Do(func() {
		// Ensure the directory exists
		dir := filepath.Dir(dataSourceName)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatalf("Failed to create directory: %v", err)
			}
		}

		// Open (or create) the database
		var err error
		db, err = sql.Open("sqlite3", dataSourceName)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		// Ensure the database connection is valid
		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Initialize the schema and initial data
		err = initSchema()
		if err != nil {
			log.Fatalf("Error initializing schema: %v", err)
		}
		err = initRoles()
		if err != nil {
			log.Fatalf("Error initializing roles: %v", err)
		}
		err = initAdminUser()
		if err != nil {
			log.Fatalf("Error initializing admin: %v", err)
		}
	})
	return err
}

func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection.
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func initSchema() error {
	// Create roles table
	roleTable := `
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT
	);`
	if _, err := db.Exec(roleTable); err != nil {
		log.Fatalf("Failed to create roles table: %v", err)
		return err
	}

	// Create users table with salt column
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		salt TEXT NOT NULL,
		role_id INTEGER,
		created_at DATETIME,
		updated_at DATETIME,
		FOREIGN KEY (role_id) REFERENCES roles(id)
	);`
	if _, err := db.Exec(userTable); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
		return err
	}
	return nil
}

func initRoles() error {
	roles := []string{"admin", "user"}

	for _, role := range roles {
		// Check if the role already exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM roles WHERE name = ?", role).Scan(&count)
		if err != nil {
			log.Fatalf("Failed to check for role %s: %v", role, err)
			return err
		}

		// Insert the role if it does not exist
		if count == 0 {
			_, err := db.Exec("INSERT INTO roles (name, description) VALUES (?, ?)", role, role)
			if err != nil {
				log.Fatalf("Failed to insert role %s: %v", role, err)
				return err
			}
		}
	}
	return nil
}

func initAdminUser() error {
	// Get admin role ID
	var roleId int
	err := db.QueryRow("SELECT id FROM roles WHERE name = ?", "admin").Scan(&roleId)
	if err != nil {
		log.Fatalf("Failed to get admin role ID: %v", err)
		return err
	}

	// Check if the admin user already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "admin").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check for admin user: %v", err)
		return err
	}

	// Insert the admin user if it does not exist
	if count == 0 {
		password := "admin123"
		salt := utils.GenerateNonce(8)
		passwordHash := utils.HashPassword(password, salt)
		_, err := db.Exec("INSERT INTO users (username, password, salt, role_id) VALUES (?, ?, ?, ?)", "admin", passwordHash, salt, roleId)
		if err != nil {
			log.Fatalf("Failed to insert admin user: %v", err)
			return err
		}
	}
	return nil
}
