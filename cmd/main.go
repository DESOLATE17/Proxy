package main

import (
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"proxy/internal/pkg/api/handler"
	"proxy/internal/pkg/api/repo"
	"proxy/internal/pkg/api/usecase"
	"proxy/internal/pkg/proxy"
	"strings"
)

func main() {
	vp := viper.New()
	if err := initConfig(vp, "config.yml"); err != nil {
		log.Printf("error initializing configs: %s\n", err.Error())
	}

	proxy := proxy.NewProxyServer()
	go func() {
		log.Fatal(proxy.ListenAndServe())
	}()

	db, err := repo.NewPostgresDB(vp.GetString("db.connection_string"))
	if err != nil {
		log.Fatal("error during connecting to postgres ", err)
	}

	repos := repo.NewRepo(db)
	services := usecase.NewUsecase(repos)
	handlers := handler.NewHandler(services)

	router := handlers.SetupRoutes()
	router.Run("127.0.0.1:8000")
}

func initConfig(vp *viper.Viper, configPath string) error {
	path := filepath.Dir(configPath)
	vp.AddConfigPath(path)
	vp.SetConfigName(strings.Split(filepath.Base(configPath), ".")[0])
	vp.SetConfigType(filepath.Ext(configPath)[1:])
	return vp.ReadInConfig()
}
