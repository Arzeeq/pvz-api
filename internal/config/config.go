package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvDev  = "dev"
	EnvProd = "prod"
	EnvTest = "test"
)

type Config struct {
	DBParam
	Env            string        `yaml:"env" env-required:"true"`
	JWTDuration    time.Duration `yaml:"jwt_duration" env-required:"true"`
	LoggerFormat   string        `yaml:"logger_format"`
	MigrationDir   string        `yaml:"migrations_dir"`
	RequestTimeout time.Duration `yaml:"request_timeout" env-default:"5s"`
	ConnectionStr  string        `yaml:"-"`
	JWTSecret      string        `yaml:"-"`
	HTTPPort       int           `yaml:"-"`
	GRPCPort       int           `yaml:"-"`
	PrometheusPort int           `yaml:"-"`
}

type DBParam struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

func (p *DBParam) GetConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", p.DBUser, p.DBPassword, p.DBHost, p.DBPort, p.DBName)
}

func MustLoad(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	httpPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		log.Fatalf("failed to read HTTP port from env: %s", err)
	}
	cfg.HTTPPort = httpPort

	grpcPort, err := strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("failed to read gRPC port from env: %s", err)
	}
	cfg.GRPCPort = grpcPort

	promPort, err := strconv.Atoi(os.Getenv("PROMETHEUS_PORT"))
	if err != nil {
		log.Fatalf("failed to read Prometheus port from env: %s", err)
	}
	cfg.PrometheusPort = promPort

	cfg.JWTSecret = os.Getenv("JWT_SECRET")

	// DBParams
	cfg.DBPassword = os.Getenv("DATABASE_PASSWORD")
	cfg.DBUser = os.Getenv("DATABASE_USER")
	cfg.DBHost = os.Getenv("DATABASE_HOST")
	cfg.DBPort = os.Getenv("DATABASE_PORT")
	cfg.DBName = os.Getenv("DATABASE_NAME")
	cfg.ConnectionStr = cfg.DBParam.GetConnStr()

	return &cfg
}
