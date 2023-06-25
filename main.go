package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	jwt_http2 "github.com/Moranilt/jwt-http2/auth"
	"github.com/Moranilt/jwt-http2/certs"
	"github.com/Moranilt/jwt-http2/clients"
	"github.com/Moranilt/jwt-http2/config"
	"github.com/Moranilt/jwt-http2/logger"
	"github.com/Moranilt/jwt-http2/middleware"
	"github.com/Moranilt/jwt-http2/server"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	log := logger.New()
	log.Info("starting application...")
	ctx, cancel := context.WithCancel(context.Background())

	env, err := config.ReadEnv()
	if err != nil {
		log.Fatalf("error while reading env: %v", err)
	}

	vaultClient, err := clients.Vault(env.Vault)
	if err != nil {
		log.Fatalf("vault client: %v", err)
	}

	// use only in local or dev modes
	if !env.Production {
		certGenerator := certs.NewKeys(vaultClient.GetClient(), env.Vault)
		err := certGenerator.StoreToVault()
		if err != nil {
			log.Fatalf("create certificates: %v", err)
		}
	}

	redisCreds, err := vaultClient.GetRedisCreds(ctx)
	if err != nil {
		log.Fatalf("vault client: %v", err)
	}

	publicCert, err := vaultClient.GetPublicCert(ctx)
	if err != nil {
		log.Fatalf("vault public cert: %v", err)
	}

	privateCert, err := vaultClient.GetPrivateCert(ctx)
	if err != nil {
		log.Fatalf("vault private cert: %v", err)
	}

	redis, err := clients.Redis(ctx, redisCreds)
	if err != nil {
		log.Fatalf("redis client: %v", err)
	}

	mw := middleware.New(log)
	service := server.New(log, redis, publicCert, privateCert)

	server := grpc.NewServer(grpc.ConnectionTimeout(10*time.Second), grpc.UnaryInterceptor(mw.UnaryInterceptor))
	jwt_http2.RegisterAuthenticationServer(server, service)
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		cancel()
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", env.Port))
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	// TODO: make watch endpoint
	// router := mux.NewRouter()
	// router.HandleFunc("/watch", )

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		<-gCtx.Done()
		server.GracefulStop()
		lis.Close()
		return fmt.Errorf("shutdown application")
	})
	g.Go(func() error {
		return server.Serve(lis)
	})

	if err := g.Wait(); err != nil {
		log.Debugf("exit with: %s", err)
	}
}
