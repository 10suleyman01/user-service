package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"learn/internal/config"
	"learn/internal/user"
	"learn/internal/user/db"
	"learn/pkg/client/mongodb"
	"learn/pkg/logging"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var logger = logging.GetLogger()

func main() {
	logger.Info("Create Router")
	router := httprouter.New()

	cfg := config.GetConfig()

	mCfg := cfg.MongoDb
	mongoDBClient, err := mongodb.NewClient(context.Background(), mCfg.Host, mCfg.Port, mCfg.Username, mCfg.Password, mCfg.Database, mCfg.AuthDb)
	if err != nil {
		panic(err)
	}
	storage := db.NewStorage(mongoDBClient, mCfg.Collection, logger)
	logger.Trace(storage)

	logger.Info("Register User Handler")
	handler := user.NewHandler(logger, storage)
	handler.Register(router)

	start(router, cfg)
}

func start(router *httprouter.Router, cfg *config.Config) {
	logger.Info("start application")

	var listener net.Listener
	var listenErr error

	if cfg.Listen.Type == "sock" {
		logger.Info("detect app path")
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("create socket")
		socketPath := path.Join(appDir, "app.sock")
		logger.Debugf("socker path: %s", socketPath)

		logger.Info("listen unix socket")
		listener, listenErr = net.Listen("unix", socketPath)
		logger.Infof("server is listening in unix %s", socketPath)
	} else {
		logger.Info("listen tcp")
		listener, listenErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		logger.Infof("server is listening in %s:%s", cfg.Listen.BindIP, cfg.Listen.Port)
	}

	if listenErr != nil {
		logger.Fatal(listenErr)
	}

	server := &http.Server{
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}
