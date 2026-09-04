package svc

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"

	_ "net/http/pprof"

	"bitbucket.org/theruister/goBase/internal/config"
	gbGrpc "bitbucket.org/theruister/goBase/internal/grpc"
	"bitbucket.org/theruister/goBase/internal/logging"
	pb "bitbucket.org/theruister/goBase/pkg/api/v1/gen/go"
	"bitbucket.org/theruister/goBase/third_party"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func getOpenAPIHandler() http.Handler {
	mime.AddExtensionType(".svg", "image/svg+xml")
	// Use subdirectory in embedded files
	subFs, err := fs.Sub(third_party.OpenAPI, "OpenAPI")
	if err != nil {
		panic("couldn't create sub filesystem: " + err.Error())
	}
	return http.FileServer(http.FS(subFs))
}

// Config ...
var Config *config.BaseConfig

func InitLogger() {
	config.InitConfiguration(".")
	Config = config.GetConfiguration()
	// Setup dynamic logging levels
	var dll logging.Level

	switch strings.ToLower(config.GetConfiguration().LogLevel) {
	case "error":
		dll = logging.Error
	case "warning":
		dll = logging.Warn
	case "info":
		dll = logging.Info
	case "trace":
		dll = logging.Trace
	default:
		dll = logging.Warn
	}

	logging.SetModuleLogLevel("all", dll)
}

func RunRPC() {
	grpcLogger := logging.GetModuleLogger(logging.LOGGER_GRPC_MODULE)

	srv := grpc.NewServer()
	pb.RegisterGoBaseServiceServer(srv, gbGrpc.New())

	addr := "0.0.0.0:" + strconv.FormatInt(int64(config.GetRpcPort()), 10)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		grpcLogger.Fatalln("Failed to listen:", err)
	}
	// Gather metrics for Prometheus
	grpc_prometheus.Register(srv)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		promAddr := "0.0.0.0:" + strconv.FormatInt(int64(config.GetPrometheusPort()), 10)
		http.ListenAndServe(promAddr, nil)
	}()

	go func() {
		cfg := config.GetConfiguration()
		http.ListenAndServe(cfg.PprofPort, nil)
	}()
	// Serve gRPC Server
	grpcLogger.Info("Serving gRPC on https://", addr)
	go func() {
		grpcLogger.Fatal(srv.Serve(lis))
	}()

	err = Run(config.GetRpcPort(), config.GetRpcGatewayPort(), false)
	if err != nil {
		grpcLogger.Error(err.Error())
	}

}

// Run ...
func Run(gRPCPort int, gRPCGatewayPort int, tlsEnabled bool) error {
	dialAddr := "dns:///0.0.0.0:" + strconv.FormatInt(int64(config.GetRpcPort()), 10)

	var dialOptTLS grpc.DialOption
	dialOptTLS = grpc.WithInsecure()

	conn, err := grpc.DialContext(
		context.Background(),
		dialAddr,
		grpc.WithBlock(),
		dialOptTLS,
	)

	if err != nil {
		Status := gbGrpc.GetServiceStatus()
		Status.GrpcError = true
		Status.GrpcMessage = fmt.Sprintf("%v", err)
		return fmt.Errorf("failed to dial server: %w", err)
	}

	gwmux := runtime.NewServeMux()
	err = pb.RegisterGoBaseServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		Status := gbGrpc.GetServiceStatus()
		Status.GrpcError = true
		Status.GrpcMessage = fmt.Sprintf("%v", err)
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	oa := getOpenAPIHandler()

	gatewayAddr := "0.0.0.0:" + strconv.FormatInt(int64(gRPCGatewayPort), 10)
	gwServer := &http.Server{
		Addr: gatewayAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				gwmux.ServeHTTP(w, r)
				return
			}
			oa.ServeHTTP(w, r)
		}),
	}
	// Empty parameters mean use the TLS Config specified with the server.
	grpcLogger := logging.GetModuleLogger(logging.LOGGER_GRPC_MODULE)
	if !tlsEnabled {
		grpcLogger.Info(fmt.Sprintf("Serving gRPC-Gateway and OpenAPI Documentation on http://%v", gatewayAddr))
		fmt.Println(fmt.Sprintf("Serving gRPC-Gateway on http://%v", gatewayAddr))
		return fmt.Errorf("serving gRPC-Gateway server: %w", gwServer.ListenAndServe())
	}

	grpcLogger.Info(fmt.Sprintf("Serving gRPC-Gateway and OpenAPI Documentation on https://%v", gatewayAddr))

	fmt.Println("Servig gRPC-Gateway on https://%v", gatewayAddr)
	return fmt.Errorf("serving gRPC-Gateway server: %w", gwServer.ListenAndServeTLS("", ""))
}

func ShutdownSvr() {
	service := "goBase App"
	fmt.Println(fmt.Sprintf("Shutting down %s service", service))
	//mq.ReportControlShutdown()
}
