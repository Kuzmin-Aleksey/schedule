package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Schedule   ScheduleConfig   `yaml:"schedule"`
	Log        LogConfig        `yaml:"log"`
	Db         DbConfig         `yaml:"db"`
	HttpServer HttpServerConfig `yaml:"http_server"`
	GrpcServer GrpcServerConfig `yaml:"grpc_server"`
}

type ScheduleConfig struct {
	NextTakingPeriod time.Duration `yaml:"next_taking_period" env:"NEXT_TAKING_PERIOD" env-default:"1h"`
	BeginDayHour     int           `yaml:"begin_day_hour" env:"BEGIN_DAY_HOUR" env-default:"8"`
	EndDayHour       int           `yaml:"end_day_hour" env:"END_DAY_HOUR" env-default:"22"`
	TimeRound        time.Duration `yaml:"time_round" env:"TIME_ROUND" env-default:"15m"`
}

type LogConfig struct {
	File  string `yaml:"file" env:"LOG_FILE" env-default:""`
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"debug"`
}

type DbConfig struct {
	Addr           string `yaml:"addr" env:"DB_ADDR" env-default:"localhost:3306"`
	User           string `yaml:"user" env:"DB_USER" env-default:"root"`
	Password       string `yaml:"password" env:"DB_PASSWORD" env-default:""`
	Schema         string `yaml:"schema" env:"DB_SCHEMA" env-default:"public"`
	ConnectTimeout int    `yaml:"connect_timeout" env:"DB_CONNECT_TIMEOUT" env-default:"10"`
}

type HttpServerConfig struct {
	Addr            string        `yaml:"addr" env:"HTTP_ADDR" env-default:"localhost:8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type GrpcServerConfig struct {
	Addr string `yaml:"addr" env:"GRPC_ADDR" env-default:"localhost:8081"`
}

func ReadConfig(path string) (*Config, error) {
	cfg := new(Config)
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
