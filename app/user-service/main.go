package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Victor-Acrani/user-api/app/user-service/api"
	v1 "github.com/Victor-Acrani/user-api/app/user-service/api/v1"
	"github.com/Victor-Acrani/user-api/extensions/logger"
	"github.com/ardanlabs/conf/v3"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const serviceName = "user-api"

type config struct {
	// service
	BuildCommit string `conf:"env:BUILD_COMMIT,default:sha256 some number"`
	BuildTime   string `conf:"env:BUILD_TIME,default:2024-09-22T13:44:51.662-0300"`
	BuildTag    string `conf:"env:BUILD_TAG,default:1.0.0"`

	// http server
	ServerReadTimeout     time.Duration `conf:"env:SERVER_READ_TIMEOUT,default:5s"`
	ServerWriteTimeout    time.Duration `conf:"env:SERVER_WRITE_TIMETOUT,default:10s"`
	ServerIdleTimeout     time.Duration `conf:"env:SERVER_iDLE_TIMEOUT,default:120s"`
	ServerShutdownTimeout time.Duration `conf:"env:SERVER_SHUTDOWN_TIMEOUT,default:20s,mask"`
	ServerAPIHost         string        `conf:"env:SERVER_API_HOST,default:0.0.0.0:3000"`
	ServerDebugHost       string        `conf:"env:SERVER_DEBUG_HOST,default:0.0.0.0:4000"`
}

func main() {
	// load .env file
	err := loadEnvFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// load env var
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create logger
	log, err := logger.New(serviceName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// print config
	out, err := conf.String(&cfg)
	if err != nil {
		fmt.Println("generating config for output: %w", err)
		os.Exit(1)
	}
	log.Infow("startup", "config", out)

	// add log complements
	mainLog := log.With(
		zap.String("build_commit", cfg.BuildCommit),
		zap.String("build_tag", cfg.BuildTag),
		zap.String("build_time", cfg.BuildTag),
		zap.Int("go_max_procs", runtime.GOMAXPROCS(0)),
		zap.Int("runtime_run_cpus", runtime.NumCPU()),
	)

	// start app
	err = run(mainLog, cfg)
	if err != nil {
		log.Error("error to start up")
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger, cfg config) error {
	// -------------------------------------------------------------------------
	// Start API Service
	log.Info("initializing V1 API support")

	// create shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// create api router
	apiRouter := api.NewRouter()
	// set api v1 routes
	apiV1 := v1.API{
		LivenessHandler:  v1.LivenessHandler(),
		ReadinessHandler: v1.ReadinessHandler(),
	}
	apiV1.Routes(apiRouter)

	// api create server
	apiServer := http.Server{
		Addr:         cfg.ServerAPIHost,
		Handler:      apiRouter,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("starting api server", zap.String("host", cfg.ServerAPIHost))
		serverErrors <- apiServer.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown
	select {
	case err := <-serverErrors:
		fmt.Println("---> serverErrors")
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		fmt.Println("---> shutdown")
		log.Info("shutdown", zap.String("status", "shutdown started"), zap.String("signal", sig.String()))
		defer log.Info("shutdown", zap.String("status", "shutdown complete"), zap.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ServerShutdownTimeout)
		defer cancel()

		if err := apiServer.Shutdown(ctx); err != nil {
			apiServer.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

// loadEnvFile loads env var from .env file.
func loadEnvFile() error {
	// filepath := "../../.env"
	filepath := ".env"

	_, err := os.Stat(filepath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	err = godotenv.Load(filepath)
	if err != nil {
		return err
	}

	return nil
}

// loadConfig loads env vars into a config struct.
func loadConfig() (config, error) {
	var cfg config
	_, err := conf.Parse("", &cfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}
