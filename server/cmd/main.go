package main

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/lquyet/distributed-lock-demo/server/config"
	"github.com/lquyet/distributed-lock-demo/server/pb"
	"github.com/lquyet/distributed-lock-demo/server/pkg/grpclib"
	log "github.com/lquyet/distributed-lock-demo/server/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	rootCmd := cobra.Command{
		Use: "server",
	}

	rootCmd.AddCommand(
		startServerCommand(),
	)
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func startServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start the server",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}
}

func startServer() {
	conf := config.Load()
	logger := config.NewLogger(conf.Log)

	grpcServer := grpc.NewServer(
		grpclib.ChainUnaryInterceptorIgnoreHealthCheck(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_zap.PayloadUnaryServerInterceptor(logger, loggingDecider),
			log.SetTraceInfoInterceptor(logger),
		),
	)

	pb.RegisterHealthCheckServiceServer(grpcServer, grpclib.NewHealthServer())
	startHTTPAndGRPCServers(conf, grpcServer)
}

func loggingDecider(_ context.Context, fullMethod string, _ interface{}) bool {
	// TODO: ignore health check methods here
	return true
}

func startHTTPAndGRPCServers(conf config.Config, grpcServer *grpc.Server) {
	fmt.Println("GRPC:", conf.Server.GRPC.ListenString())
	fmt.Println("HTTP:", conf.Server.HTTP.ListenString())

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	ctx := context.Background()
	grpcHost := conf.Server.GRPC.String()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	registerGRPCGateway(ctx, mux, grpcHost, opts)

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.Handle("/", mux)

	httpServer := &http.Server{
		Addr:    conf.Server.HTTP.ListenString(),
		Handler: allowCORS(httpMux),
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
		fmt.Println("Shutdown HTTP server successfully")
	}()

	go func() {
		defer wg.Done()

		listener, err := net.Listen("tcp", conf.Server.GRPC.ListenString())
		if err != nil {
			panic(err)
		}

		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
		fmt.Println("Shutdown gRPC server successfully")
	}()

	sigCh := make(chan bool, 1)

	// --------------------------------
	// Graceful Shutdown
	// --------------------------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sigCh <- true

	grpcServer.GracefulStop()
	err := httpServer.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	wg.Wait()
}

func registerGRPCGateway(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) {
	_ = pb.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// preflightHandler adds the necessary headers in order to serve
// CORS from any origin using the methods "GET", "HEAD", "POST", "PUT", "DELETE"
// We insist, don't do this without consideration in production systems.
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
}
