package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	vfsOs "github.com/c2fo/vfs/v6/backend/os"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	crt "gitlab.com/DzmitryYafremenka/golang-united-school-certs"
	"gitlab.com/DzmitryYafremenka/golang-united-school-certs/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpc and rest on same port
func grpcHandler(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func main() {
	serverHost := ":8080"
	httpHost := "http://localhost:8080/"
	gotenberg := "http://localhost:3000"
	connString := "postgres://user:password@localhost:5432/registry"
	outPath := "./tmp/e2e/demo/"

	_, docker := os.LookupEnv("DOCKER")
	if docker {
		gotenberg = "http://gotenberg:3000"
		connString = "postgres://user:password@db:5432/registry"
		outPath = "./e2e/demo/"
	}

	dr, err := crt.NewDirectRegistry(connString)
	if err != nil {
		log.Fatalf("Failed to create DirectRegistry: %v", err)
	}
	r, err := crt.NewCachedRegistry(dr)
	if err != nil {
		log.Fatalf("Failed to create CachedRegistry: %v", err)
	}

	path, err := filepath.Abs(outPath + time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatalf("Failed to resolve absolute path to ./tmp: %v", err)
	}
	s, err := crt.NewVfsStorage("", path+"/", vfsOs.Scheme, nil)
	if err != nil {
		log.Fatalf("Failed to create VfsStorage: %v", err)
	}
	log.Println("Storage pointing to:", path)

	t := crt.NewGotenbergTemplater(gotenberg)
	server := crt.NewCertsServer(r, s, t, httpHost)

	grpcServer := grpc.NewServer()
	api.RegisterCertsServiceServer(grpcServer, server)

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = api.RegisterCertsServiceHandlerFromEndpoint(context.Background(), mux, serverHost, opts)
	if err != nil {
		log.Fatalf("Failed to register service handler: %v", err)
	}

	lis, err := net.Listen("tcp", serverHost)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Serving on:", serverHost)
	log.Fatal(http.Serve(lis, grpcHandler(grpcServer, mux)))
}
