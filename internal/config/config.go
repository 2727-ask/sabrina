package config

import (
	"fmt"
	"net/url"
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

	host := "localhost:5672" // local port forwarding

	if env == "production" {
		host = get("RABBITMQ_HOST", "rabbitmq.jobstar.svc.cluster.local:5672")
	}

	user := get("RABBITMQ_USER", "user")
	pass := get("RABBITMQ_PASS", "guest")

	return fmt.Sprintf(
		"amqp://%s:%s@%s/",
		url.QueryEscape(user),
		url.QueryEscape(pass),
		host,
	)
}

func get(k, def string) string {
	v := os.Getenv(k)
	if v != "" {
		return v
	}
	return def
}
