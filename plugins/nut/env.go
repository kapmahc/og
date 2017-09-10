package nut

import (
	"fmt"

	"github.com/spf13/viper"
)

// IsProduction production mode?
func IsProduction() bool {
	return viper.GetString("env") == "production"
}

// DataSource datasource url
func DataSource() string {
	//"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s"
	args := ""
	for k, v := range viper.GetStringMapString("database.args") {
		args += fmt.Sprintf(" %s=%s ", k, v)
	}
	return args

	//"postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full")
	// return fmt.Sprintf(
	// 	"%s://%s:%s@%s:%d/%s?sslmode=%s",
	// 	viper.GetString("database.driver"),
	// 	viper.GetString("database.args.user"),
	// 	viper.GetString("database.args.password"),
	// 	viper.GetString("database.args.host"),
	// 	viper.GetInt("database.args.port"),
	// 	viper.GetString("database.args.dbname"),
	// 	viper.GetString("database.args.sslmode"),
	// )
}

func init() {

	viper.SetEnvPrefix("og")
	viper.BindEnv("env")

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetDefault("env", "development")
	viper.SetDefault("s3", map[string]interface{}{
		"endpoint":   "http://127.0.0.1:9000",
		"access_key": "guest",
		"secret_key": "change-me",
	})
	viper.SetDefault("redis", map[string]interface{}{
		"host": "localhost",
		"port": 6379,
		"db":   8,
	})

	viper.SetDefault("rabbitmq", map[string]interface{}{
		"user":     "guest",
		"password": "guest",
		"host":     "localhost",
		"port":     "5672",
		"virtual":  "og-dev",
	})

	viper.SetDefault("database", map[string]interface{}{
		"driver": "postgres",
		"args": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"user":     "postgres",
			"password": "",
			"dbname":   "og_dev",
			"sslmode":  "disable",
		},
		"pool": map[string]int{
			"max_open": 180,
			"max_idle": 6,
		},
	})

	viper.SetDefault("server", map[string]interface{}{
		"port":     8080,
		"ssl":      false,
		"name":     "change-me.com",
		"frontend": []string{"http://localhost:3000"},
		"backend":  "http://localhost:8080",
	})

	viper.SetDefault("secrets", map[string]interface{}{
		"jwt":  Random(32),
		"aes":  Random(32),
		"hmac": Random(32),
	})

	viper.SetDefault("elasticsearch", map[string]interface{}{
		"host": "localhost",
		"port": 9200,
	})

}
