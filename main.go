package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Moranilt/jwt-http2/certs"
	"github.com/Moranilt/jwt-http2/clients"
	"github.com/Moranilt/jwt-http2/config"
	"github.com/Moranilt/jwt-http2/logger"
	"github.com/Moranilt/jwt-http2/middleware"
	"github.com/Moranilt/jwt-http2/server"
	grpc_transport "github.com/Moranilt/jwt-http2/transport/grpc"
	http_transport "github.com/Moranilt/jwt-http2/transport/http"
	"golang.org/x/sync/errgroup"
)

func main() {
	log := logger.New()
	log.Info("starting application...")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		cancel()
	}()

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

	consulClient, err := clients.Consul(ctx, env.Consul)
	if err != nil {
		log.Fatalf("consul client: %v", err)
	}

	mainConfig := config.New(log)
	err = mainConfig.ReadConsul(ctx, env.Consul.Key(), consulClient)
	if err != nil {
		log.Fatal("read from consul: ", err)
	}

	serverREST := http_transport.New(fmt.Sprintf(":%s", env.PortREST), log, mainConfig, env.Consul.Key())
	mw := middleware.New(log)
	server := server.New(log, mainConfig.App, redis, publicCert, privateCert)
	serverGRPC := grpc_transport.New(server, mw)
	lis, err := serverGRPC.MakeListener(env.PortGRPC)
	if err != nil {
		log.Fatal(err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		<-gCtx.Done()
		serverGRPC.GracefulStop()
		lis.Close()
		serverREST.Shutdown(context.Background())
		return fmt.Errorf("shutdown application")
	})

	g.Go(func() error {
		return serverGRPC.Serve(lis)
	})

	g.Go(func() error {
		return serverREST.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Debugf("exit with: %s", err)
	}
}
