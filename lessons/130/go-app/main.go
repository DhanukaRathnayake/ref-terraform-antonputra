package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/argon2"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type user struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	r := gin.New()
	r.POST("/users", createUser)
	r.GET("/metrics", prometheusHandler())
	r.Run()
}

var generateHashDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "generate_hash_duration_seconds",
	Help:    "Duration to generate argon2 hash for the user.",
	Buckets: []float64{0.05, 0.06, 0.07, 0.08, 0.09, 0.1, 0.11, 0.12, 0.13, 0.14},
})

var saveUserDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "save_user_duration_seconds",
	Help:    "Duration to save user into the database.",
	Buckets: []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09, 0.1},
})

func init() {
	prometheus.MustRegister(generateHashDuration, saveUserDuration)
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// createUser creates a new user with generated password and stores it in the database.
func createUser(c *gin.Context) {
	var newUser user

	if err := c.BindJSON(&newUser); err != nil {
		http.Error(c.Writer, "failed to create user", 500)
		log.Fatal(err)
	}

	p := &params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}

	encodedHash, err := generateFromPassword(newUser.Password, p)
	if err != nil {
		http.Error(c.Writer, "failed to create user", 500)
		log.Fatal(err)
	}

	err = saveUser(newUser.Email, encodedHash)
	if err != nil {
		http.Error(c.Writer, "failed to create user", 500)
		log.Fatalln(err)
	}

	c.String(http.StatusCreated, "User created.")
}

// generateFromPassword generates a hash from the provided password.
func generateFromPassword(password string, p *params) (encodedHash string, err error) {
	start := time.Now()
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", fmt.Errorf("generateRandomBytes: %v", err)
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	generateHashDuration.Observe(float64(time.Since(start).Seconds()))
	return encodedHash, nil
}

// generateRandomBytes generates a random byte slice to use as salt.
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("rand.Read: %v", err)
	}

	return b, nil
}

// saveUser saves the user into the database.
func saveUser(email string, hash string) error {
	start := time.Now()
	connStr := "user=app password=devops123 dbname=lesson_130 host=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("sql.Open: %v", err)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO go_users(email, password_hash) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("db.Prepare: %v", err)
	}
	_, err = stmt.Exec(email, hash)
	if err != nil {
		return fmt.Errorf("stmt.Exec: %v", err)
	}
	saveUserDuration.Observe(float64(time.Since(start).Seconds()))
	return nil
}
