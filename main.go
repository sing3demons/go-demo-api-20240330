package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return err
	} else {
		for _, key := range viper.AllKeys() {
			if strings.HasPrefix(key, "app.") {
				newKey := strings.ToUpper(strings.TrimPrefix(key, "app."))
				newKey = strings.ReplaceAll(newKey, ".", "_")
				value := fmt.Sprintf("%v", viper.Get(key))
				os.Setenv(newKey, value)
			}
		}
	}
	return nil
}

func initTimeZone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}

	time.Local = ict
}

func initLogger() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				a.Key = "timestamp"
			}

			return a
		},
	})))
}

func init() {
	initLogger()
	initTimeZone()
	if err := initConfig(); err != nil {
		panic(err)
	}

}

type m map[string]any

func main() {
	router := NewRouter()

	port := os.Getenv("PORT")
	appName := os.Getenv("APP_NAME")

	router.GET("/", func(w http.ResponseWriter, req *http.Request) {
		sessionId := Session(req.Context())

		json.NewEncoder(w).Encode(m{
			"message": "Hello, World!",
			"session": sessionId,
		})
	})

	router.StartHTTP(appName, port)
}
