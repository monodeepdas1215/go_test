package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/monodeepdas1215/go_test/core"
	data_store "github.com/monodeepdas1215/go_test/core/data-store"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// initializing the application database connection
	data_store.AppDb = data_store.GetNewPgDatabaseConnection()

	// initializing all the middlewares if any
	middlewares := make([]gin.HandlerFunc, 0)
	App := core.NewApp("localhost", "8080", middlewares)
	App.StartApplication()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("please wait while server is shutting down ...")

	ctx, cancel :=  context.WithTimeout(context.Background(), 5*time.Second)
	App.ShutdownServer(ctx, cancel)

	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		data_store.AppDb.Disconnect()
	}
	log.Println("server shut down gracefully")
}
