package config

type Config struct {
	App      AppConfig
	GRPC     GRPCConfig
	Postgres PostgresConfig
}

// =======================
// App (identity)
// =======================

type AppConfig struct {
	Name string `env:"APP_NAME" env-required:"true"`
	Env  string `env:"APP_ENV" env-default:"local"`
}

// =======================
// GRPC Runtime
// =======================

type GRPCConfig struct {
	Port       int  `env:"GRPC_PORT" env-default:"50051"`
	Reflection bool `env:"REFLECTION" env-default:"false"`
}

// =======================
// Database
// =======================

type PostgresConfig struct {
	Host     string `env:"DB_HOST" env-required:"true"`
	Port     int    `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	DBName   string `env:"DB_NAME" env-required:"true"`
	SSLMode  string `env:"DB_SSLMODE" env-default:"disable"`

	MaxOpenConns    int `env:"MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns    int `env:"MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime int `env:"CONN_MAX_LIFETIME" env-default:"5"`
}
