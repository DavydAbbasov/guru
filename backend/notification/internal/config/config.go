package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Server    *ServerConfig    `mapstructure:"server"`
	Database  *DatabaseConfig  `mapstructure:"database" validate:"required"`
	Kafka     *KafkaConfig     `mapstructure:"kafka"    validate:"required"`
	Logger    *LoggerConfig    `mapstructure:"logger"`
	Tracer    *TracerConfig    `mapstructure:"tracer"`
	Metrics   *MetricsConfig   `mapstructure:"metrics"`
	Profiling *ProfilingConfig `mapstructure:"profiling"`
}

type DatabaseConfig struct {
	Host string `mapstructure:"host" validate:"required"`
	Port int    `mapstructure:"port" validate:"required,gt=0"`
	User string `mapstructure:"user" validate:"required"`
	Pass string `mapstructure:"pass" validate:"required"`
	Name string `mapstructure:"name" validate:"required"`
}

type ProfilingConfig struct {
	CPU    bool   `mapstructure:"cpu"`
	Memory bool   `mapstructure:"memory"`
	Path   string `mapstructure:"path"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port" validate:"required,gt=0"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers" validate:"required,min=1,dive,required"`
	Topic   string   `mapstructure:"topic"   validate:"required"`
	GroupID string   `mapstructure:"groupId" validate:"required"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

type TracerConfig struct {
	Disabled    bool   `mapstructure:"disabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"serviceName"`
}

type MetricsConfig struct {
	Namespace string `mapstructure:"namespace"`
}

func New() (*Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8081)
	v.SetDefault("logger.level", "info")
	v.SetDefault("tracer.disabled", true)
	v.SetDefault("metrics.namespace", "notification")

	// CFG_KAFKA_BROKERS → kafka.brokers
	v.SetEnvPrefix("CFG")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			var errStr strings.Builder
			for _, e := range validationErrs {
				fmt.Fprintf(&errStr, "validation failed for field '%s': %s\n", e.Field(), e.ActualTag())
			}
			return nil, fmt.Errorf("config validation failed: %s", errStr.String())
		}
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return cfg, nil
}
