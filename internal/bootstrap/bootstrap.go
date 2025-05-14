package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/runsystemid/golog"
	"github.com/runsystemid/gontainer"
	"gitlab.com/ayaka/config"
	"gitlab.com/ayaka/internal/adapter/rest"
	"gitlab.com/ayaka/internal/pkg/email"
	"gitlab.com/ayaka/internal/pkg/validator"
)

var appContainer = gontainer.New()

func Run(conf *config.Config) {
	appContainer.RegisterService("config", conf)
	appContainer.RegisterService("mail", new(email.MailSMTP))

	// Initialize struct validator
	appContainer.RegisterService("validator", validator.NewGoValidator())

	bootstrapContext, cancel := context.WithCancel(context.Background())
	golog.Info(bootstrapContext, "Serving...")

	// Register adapter
	RegisterDatabase()
	RegisterCache()
	RegisterRest()
	// RegisterToggleService()
	RegisterRepository()

	// Register application
	RegisterService()
	RegisterApi()

	// Register another handler
	RegisterHandler()

	// Startup the container
	if err := appContainer.Ready(); err != nil {
		golog.Panic(bootstrapContext, "Failed to populate service", err)
	}

	// Start server
	fiberApp := appContainer.GetServiceOrNil("fiber").(*rest.Fiber)
	errs := make(chan error, 2)
	go func() {
		golog.Info(bootstrapContext, fmt.Sprintf("Listening on port :%d", conf.Http.Port))
		errs <- fiberApp.Listen(fmt.Sprintf(":%d", conf.Http.Port))
	}()

	golog.Info(bootstrapContext, "Your app started")
	printLogo(conf.Http.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		golog.Info(bootstrapContext, "Signal termination received")
		cancel()
	}()

	<-bootstrapContext.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	golog.Info(shutdownCtx, "Cleaning up resources...")

	appContainer.Shutdown()

	golog.Info(shutdownCtx, "Bye")
}

func printLogo(port int) {
	fmt.Printf(`
     ░█▀▀█ ░█──░█ ░█▀▀█ ░█─▄▀ ░█▀▀█ 
     ▒█▄▄█ ░█▄▄▄█ ▒█▄▄█ ░█▀▄─ ▒█▄▄█ 
     ▒█─▒█ ──░█── ▒█─▒█ ░█─░█ ▒█─▒█ 
    Running on port %d
`+"\n", port)
}
