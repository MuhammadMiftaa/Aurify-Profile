package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type (
	Server struct {
		Mode     string `env:"MODE"`
		HTTPPort string `env:"HTTP_PORT"`
		GRPCPort string `env:"GRPC_PORT"`
	}

	Database struct {
		DBHost     string `env:"DB_HOST"`
		DBPort     string `env:"DB_PORT"`
		DBUser     string `env:"DB_USER"`
		DBPassword string `env:"DB_PASSWORD"`
		DBName     string `env:"DB_NAME"`
	}

	Minio struct {
		Host        string `env:"MINIO_HOST"`
		AccessKey   string `env:"MINIO_ROOT_USER"`
		SecretKey   string `env:"MINIO_ROOT_PASSWORD"`
		MaxOpenConn int    `env:"MINIO_MAX_OPEN_CONN"`
		UseSSL      int    `env:"MINIO_USE_SSL"`
		BucketName  string `env:"MINIO_BUCKET_NAME"`
	}

	Config struct {
		Server   Server
		Database Database
		Minio    Minio
	}
)

var Cfg Config

func LoadNative() ([]string, error) {
	var ok bool
	var missing []string

	if _, err := os.Stat("/app/.env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	// ! Load Server configuration ____________________________
	if Cfg.Server.Mode, ok = os.LookupEnv("MODE"); !ok {
		missing = append(missing, "MODE env is not set")
	}
	if Cfg.Server.HTTPPort, ok = os.LookupEnv("HTTP_PORT"); !ok {
		missing = append(missing, "HTTP_PORT env is not set")
	}
	if Cfg.Server.GRPCPort, ok = os.LookupEnv("GRPC_PORT"); !ok {
		missing = append(missing, "GRPC_PORT env is not set")
	}
	// ! ______________________________________________________

	// ! Load Database configuration __________________________
	if Cfg.Database.DBUser, ok = os.LookupEnv("DB_USER"); !ok {
		missing = append(missing, "DB_USER env is not set")
	}
	if Cfg.Database.DBHost, ok = os.LookupEnv("DB_HOST"); !ok {
		missing = append(missing, "DB_HOST env is not set")
	}
	if Cfg.Database.DBPort, ok = os.LookupEnv("DB_PORT"); !ok {
		missing = append(missing, "DB_PORT env is not set")
	}
	if Cfg.Database.DBName, ok = os.LookupEnv("DB_NAME"); !ok {
		missing = append(missing, "DB_NAME env is not set")
	}
	if Cfg.Database.DBPassword, ok = os.LookupEnv("DB_PASSWORD"); !ok {
		missing = append(missing, "DB_PASSWORD env is not set")
	}
	// ! ______________________________________________________

	// ! Load MinIO configuration _____________________________
	if Cfg.Minio.Host, ok = os.LookupEnv("MINIO_HOST"); !ok {
		missing = append(missing, "MINIO_HOST env is not set")
	}
	if Cfg.Minio.AccessKey, ok = os.LookupEnv("MINIO_ROOT_USER"); !ok {
		missing = append(missing, "MINIO_ROOT_USER env is not set")
	}
	if Cfg.Minio.SecretKey, ok = os.LookupEnv("MINIO_ROOT_PASSWORD"); !ok {
		missing = append(missing, "MINIO_ROOT_PASSWORD env is not set")
	}
	if maxConn, ok := os.LookupEnv("MINIO_MAX_OPEN_CONN"); ok {
		Cfg.Minio.MaxOpenConn, _ = strconv.Atoi(maxConn)
	}
	if useSSL, ok := os.LookupEnv("MINIO_USE_SSL"); ok {
		Cfg.Minio.UseSSL, _ = strconv.Atoi(useSSL)
	}
	if Cfg.Minio.BucketName, ok = os.LookupEnv("MINIO_BUCKET_NAME"); !ok {
		Cfg.Minio.BucketName = "profile-photos" // default bucket name
	}
	// ! ______________________________________________________

	return missing, nil
}

func LoadByViper() ([]string, error) {
	var missing []string

	viper.SetConfigFile("config.json")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Server configuration
	if Cfg.Server.Mode = viper.GetString("server.mode"); Cfg.Server.Mode == "" {
		missing = append(missing, "server.mode not set in config.json")
	}
	if Cfg.Server.HTTPPort = viper.GetString("server.http_port"); Cfg.Server.HTTPPort == "" {
		missing = append(missing, "server.http_port not set in config.json")
	}
	if Cfg.Server.GRPCPort = viper.GetString("server.grpc_port"); Cfg.Server.GRPCPort == "" {
		missing = append(missing, "server.grpc_port not set in config.json")
	}

	// Database configuration
	if Cfg.Database.DBUser = viper.GetString("database.user"); Cfg.Database.DBUser == "" {
		missing = append(missing, "database.user not set in config.json")
	}
	if Cfg.Database.DBHost = viper.GetString("database.host"); Cfg.Database.DBHost == "" {
		missing = append(missing, "database.host not set in config.json")
	}
	if Cfg.Database.DBPort = viper.GetString("database.port"); Cfg.Database.DBPort == "" {
		missing = append(missing, "database.port not set in config.json")
	}
	if Cfg.Database.DBName = viper.GetString("database.name"); Cfg.Database.DBName == "" {
		missing = append(missing, "database.name not set in config.json")
	}
	if Cfg.Database.DBPassword = viper.GetString("database.password"); Cfg.Database.DBPassword == "" {
		missing = append(missing, "database.password not set in config.json")
	}

	// MinIO configuration
	if Cfg.Minio.Host = viper.GetString("minio.host"); Cfg.Minio.Host == "" {
		missing = append(missing, "minio.host not set in config.json")
	}
	if Cfg.Minio.AccessKey = viper.GetString("minio.access_key"); Cfg.Minio.AccessKey == "" {
		missing = append(missing, "minio.access_key not set in config.json")
	}
	if Cfg.Minio.SecretKey = viper.GetString("minio.secret_key"); Cfg.Minio.SecretKey == "" {
		missing = append(missing, "minio.secret_key not set in config.json")
	}
	Cfg.Minio.MaxOpenConn = viper.GetInt("minio.max_open_conn")
	Cfg.Minio.UseSSL = viper.GetInt("minio.use_ssl")
	if Cfg.Minio.BucketName = viper.GetString("minio.bucket_name"); Cfg.Minio.BucketName == "" {
		Cfg.Minio.BucketName = "profile-photos"
	}

	return missing, nil
}
