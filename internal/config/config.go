package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	AMQPURL     string
}

func Load() Config {
	_ = godotenv.Load()
	env := get("APP_ENV", "local")

	return Config{
		Env:         env,
		Port:        get("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AMQPURL:     amqpURL(env),
	}
}

func amqpURL(env string) string {
	if v := os.Getenv("AMQP_URL"); v != "" {
		return v
	}
	user := get("RABBITMQ_USER", "guest")
	pass := get("RABBITMQ_PASS", "guest")

	def := get("RABBITMQ_HOST", "localhost:58167") // local: port-forward
	if env == "production" {
		def = "rabbitmq.<namespace>.svc.cluster.local:5672"
	}
	host := get("RABBITMQ_HOST", def)

	return fmt.Sprintf("amqp://%s:%s@%s/", user, pass, host)
}

func get(k, def string) string {
	v := os.Getenv(k)
	if v != "" {
		return v
	}
	return def
}
