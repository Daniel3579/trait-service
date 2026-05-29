package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"trait-service/db"
	"trait-service/handlers"
	"trait-service/logger"
	"trait-service/utils"

	trait_pb "github.com/Daniel3579/trait-service-sdk/gen"
	"go.uber.org/zap"

	mid "trait-service/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// init logger
	if err := logger.Init(true); err != nil {
		panic(err)
	}
	defer logger.Sync()

	// ——————————————————————————————————————————————————————————————————————————

	err := utils.LoadEnv()
	if err != nil {
		logger.Log.Fatal("Failed to load environment variables", zap.Error(err))
	}
	logger.Log.Info("Environment variables loaded successfully")

	// ——————————————————————————————————————————————————————————————————————————

	err = db.ConnectDB("DATABASE_URL")
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			logger.Log.Error("Failed to close database connection", zap.Error(err))
		}
	}()
	logger.Log.Info("Database connection established")

	// ——————————————————————————————————————————————————————————————————————————

	certFile := os.Getenv("TRAIT_SERVICE_CERT_FILE")
	if certFile == "" {
		logger.Log.Fatal("TRAIT_SERVICE_CERT_FILE environment variable is not set")
	}

	keyFile := os.Getenv("TRAIT_SERVICE_KEY_FILE")
	if keyFile == "" {
		logger.Log.Fatal("TRAIT_SERVICE_KEY_FILE environment variable is not set")
	}

	// Загружаем TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		logger.Log.Fatal("Failed to load TLS credentials: %v", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			mid.ValidateMiddleware,
			mid.UnaryMetricsInterceptor(),
		),
	)

	// Регистрация сервиса
	server := &handlers.Server{}
	trait_pb.RegisterTraitServiceServer(grpcServer, server)

	// Создаем сетевой слушатель
	var port string = os.Getenv("TRAIT_SERVICE_GRPC_PORT")
	if port == "" {
		logger.Log.Fatal("TRAIT_SERVICE_GRPC_PORT environment variable is not set")
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Fatal("Failed to listen on port", zap.Error(err))
	}
	logger.Log.Info("Trait gRPC server is running", zap.String("port", port))

	// ——————————————————————————————————————————————————————————————————————————

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Log.Info("Shutting down gRPC server")
		grpcServer.GracefulStop()
		logger.Log.Info("gRPC server stopped gracefully")
	}()

	go func() {
		var rest_port string = os.Getenv("TRAIT_SERVICE_REST_PORT")
		if rest_port == "" {
			logger.Log.Fatal("TRAIT_SERVICE_REST_PORT environment variable is not set")
		}

		// Загружаем TLS credentials
		tlsConfig, err := utils.LoadTLSConfig()
		tlsCreds := credentials.NewTLS(tlsConfig)
		if err != nil {
			logger.Log.Error("Failed to load TLS credentials", zap.Error(err))
		}

		// Создаем подключение с TLS
		address := "localhost" + port
		conn, err := grpc.NewClient(
			address,
			grpc.WithTransportCredentials(tlsCreds),
		)

		if err != nil {
			logger.Log.Error("gRPC dial auth server failed", zap.Error(err), zap.String("addr", address))
		}
		defer conn.Close()

		client := trait_pb.NewTraitServiceClient(conn)

		hs := &handlers.HttpServer{GrpcSrv: client}
		http.HandleFunc("/update", handlers.EnableCORS(hs.Update))
		http.HandleFunc("/read", handlers.EnableCORS(hs.Read))

		server := &http.Server{
			Addr:      rest_port,
			TLSConfig: tlsConfig,
		}

		logger.Log.Info("Trait REST server is running", zap.String("port", rest_port))
		if err := server.ListenAndServeTLS("", ""); err != nil {
			logger.Log.Fatal("Failed to serve", zap.Error(err))
		}

	}()

	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("Failed to serve", zap.Error(err))
	}
}
