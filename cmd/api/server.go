//Filename: cmd/api/server.go

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	//create our HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	//The Shutdown function should return its error to this channel. channel used to communicate with serve()
	shutdownError := make(chan error)

	//start a background Goroutine
	go func() {
		//Creates a quit/exit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)
		//Listen for SIGNAL AND SIGTERM signals
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		//Block until a signal is recieved
		s := <-quit
		//Log a message
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		//shutdown
		//Create a context with a 20-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		//call the Shutdown() function
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Log a message about the goroutines
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})
		app.wg.Wait()
		//once nil it execute next line
		shutdownError <- nil

	}()

	//start our server
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	//check if the shutdown process has been initiated
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	//Block for notification from the Shutdown() function
	err = <-shutdownError
	if err != nil {
		return err
	}

	//Graceful shutdown was successful
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
