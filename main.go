package main

import (
	"context"
	"fmt"
	"headscale-panel/common"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/middleware"
	"headscale-panel/repository"
	"headscale-panel/routes"
	task "headscale-panel/tasks"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// load config to global config struct
	config.InitConfig()

	// init logger
	log.InitLogger()

	// init database
	common.InitDB()

	// Initialising the casbin policy manager
	common.InitCasbinEnforcer()

	// Initialization of Validator data validation
	common.InitValidate()

	// init database data
	common.InitData()

	// load headscale config to global config struct when the mode is not multi
	config.InitHeadscaleConfig()

	// init headscale config info for headscale-panel
	common.InitHeadscale()

	// init tasks if mode is not multi it will be init by init_tasks.go
	tk, err := task.InitTasks()
	if err != nil {
		log.Log.Error(err)
		panic(err)
	}

	// Instead of sending the logs to rabbitmq or kafka, the operation logging middleware sends the logs to a channel
	// Here, three goroutines are enabled to handle the channels and log to the database
	logRepository := repository.NewOperationLogRepository()
	for i := 0; i < 3; i++ {
		go logRepository.SaveOperationLogChannel(middleware.OperationLogChan)
	}

	// register all routes
	r := routes.InitRoutes()

	srv := &http.Server{
		Addr:    config.Conf.System.ListenAddr,
		Handler: r,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Log.Info(fmt.Sprintf("Server is running at %s/%s", config.Conf.System.ListenAddr, config.Conf.System.UrlPathPrefix))

	// when enable the oidc provider, headscale need to connect oidc provider, so the task init must after server start
	if err = tk.Start(); err != nil {
		log.Log.Error(err)
		panic(err)
	}

	// init grpc to connect headscale
	// need to sleep some second to wait the headscale ready
	task.InitGRPC()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Log.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Log.Fatal("Server forced to shutdown:", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tk.Stop(ctx)

	// Wait for existing connections to finish
	log.Log.Info("Server exiting!")
}
