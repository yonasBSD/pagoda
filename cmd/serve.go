/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mikestefanello/pagoda/pkg/handlers"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/mikestefanello/pagoda/pkg/tasks"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
    // Start a new container
    c := services.NewContainer()
    defer func() {
      if err := c.Shutdown(); err != nil {
        log.Fatal(err)
      }
    }()

    // Build the router
    if err := handlers.BuildRouter(c); err != nil {
      log.Fatalf("failed to build the router: %v", err)
    }

    // Start the server
    go func() {
      port, _ := cmd.Flags().GetInt("port")

      if port == 3000 && c.Config.HTTP.Port != 3000 {
        port = int(c.Config.HTTP.Port)
      }

      srv := http.Server{
        Addr:         fmt.Sprintf("%s:%d", c.Config.HTTP.Hostname, uint16(port)),
        Handler:      c.Web,
        ReadTimeout:  c.Config.HTTP.ReadTimeout,
        WriteTimeout: c.Config.HTTP.WriteTimeout,
        IdleTimeout:  c.Config.HTTP.IdleTimeout,
      }

      if c.Config.HTTP.TLS.Enabled {
        certs, err := tls.LoadX509KeyPair(c.Config.HTTP.TLS.Certificate, c.Config.HTTP.TLS.Key)
        if err != nil {
          log.Fatalf("cannot load TLS certificate: %v", err)
        }

        srv.TLSConfig = &tls.Config{
          Certificates: []tls.Certificate{certs},
        }
      }

      if err := c.Web.StartServer(&srv); errors.Is(err, http.ErrServerClosed) {
        log.Fatalf("shutting down the server: %v", err)
      }
    }()

    // Register all task queues
    tasks.Register(c)

    // Start the task runner to execute queued tasks
    c.Tasks.Start(context.Background())

    // Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    signal.Notify(quit, os.Kill)
    <-quit

    // Shutdown both the task runner and web server
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    wg := sync.WaitGroup{}
    wg.Add(2)

    go func() {
      defer wg.Done()
      c.Tasks.Stop(ctx)
    }()

    go func() {
      defer wg.Done()
      if err := c.Web.Shutdown(ctx); err != nil {
        log.Fatal(err)
      }
    }()

    wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	serveCmd.Flags().IntP("port", "p", 3000, "Port to use for HTTP connections")
}
