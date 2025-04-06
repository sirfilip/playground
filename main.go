package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"

	"sirfilip/playground/task_manager"
)

// TODO fix flag
// type Address string
//
//	func (addr *Address) String() string {
//		return string(*addr)
//	}
//
//	func (addr *Address) Set(value string) error {
//		fmt.Printf("Parsing address: %q\n", value)
//		if value == "" {
//			*addr = Address(":3000")
//			return nil
//		}
//
//		chars := []rune(value)
//		err := errors.New("invalid address")
//		if len(chars) < 3 {
//			return err
//		}
//		if chars[0] != ':' {
//			return err
//		}
//
//		for i := 1; i < len(chars); i++ {
//			if !unicode.IsDigit(chars[i]) {
//				return err
//			}
//		}
//		*addr = Address(value)
//		return nil
//	}

type config struct {
	Addr string
}

// TODO setup env termination
func main() {
	cfg := config{}

	flag.StringVar(&cfg.Addr, "addr", ":3000", "Server http address")
	flag.Parse()

	if err := run(&cfg); err != nil {
		log.Fatal(err)
	}
}

var healthOk = []byte("ok")

func run(cfg *config) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write(healthOk)
	})

	db, err := sql.Open("sqlite3", "./task_manager/db.sqlite3")
	if err != nil {
		return fmt.Errorf("connecting to task_manager sqlite db: %w", err)
	}
	defer db.Close()

	api := task_manager.NewApi(task_manager.NewSqliteRepo(db), logger)

	mux.Handle("/api/v1/task_manager", api)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	g := new(errgroup.Group)
	g.Go(func() error {
		logger.Info(fmt.Sprintf("Starting http server on address: %v", server.Addr))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server listen error: %w", err)
		}
		return nil
	})

	<-ctx.Done()
	logger.Info("Shutting down HTTP server gracefully...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("Server shutdown error: %w", err)
	}

	logger.Info("Server shutdown success")

	return g.Wait()
}
