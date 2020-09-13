package core

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/monodeepdas1215/go_test/core/controllers"
	"github.com/monodeepdas1215/go_test/core/logger"
	"net/http"
)

type App struct {
	host, port	string
	server	*http.Server
}

func (app *App) initializeRoutes(middlewares []gin.HandlerFunc) {

	router := gin.Default()

	// adding all middlewares passed
	for i, _ := range middlewares {
		router.Use(middlewares[i])
	}

	router.PUT("/transactionservice/transaction/:transaction_id", controllers.AddTransactionController)
	router.GET("/transactionservice/:queryType/:val", controllers.GetTransactionTypesController)
	//router.GET("/transactionservice/sum/:parent_id", controllers.GetTransactionSumController)

	// initializing the server
	url := app.host + ":" + app.port

	app.server = &http.Server{
		Addr: url,
		Handler: router,
	}
}

func (app *App) StartApplication() {

	logger.Logger.Infoln("starting application...")
	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Errorln("error occurred while starting server: ", err)
		}
	}()
}

func (app *App) ShutdownServer(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		logger.Logger.Errorln("Error occurred while shutting down server : ", err)
	}
}

func NewApp(host, port string, middlewares []gin.HandlerFunc) *App {

	app := &App{
		host: host,
		port: port,
		server: nil,
	}
	app.initializeRoutes(middlewares)
	return app
}