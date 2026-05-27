package auth

import (
	"context"
	"fmt"
	"os"
	"time"
	"trait-service/logger"
	"trait-service/utils"

	auth_pb "github.com/Daniel3579/auth-service-sdk/gen"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func RequestValidate(accessToken string) (string, int, error, codes.Code) {
	// Подключаемся к gRPC серверу
	var address string = os.Getenv("AUTH_SERVER_GRPC_PORT")
	if address == "" {
		logger.Log.Error("AUTH_SERVER_GRPC_PORT not set")
		return "", -1, fmt.Errorf("AUTH_SERVER_GRPC_PORT environment variable is not set"), codes.InvalidArgument
	}

	// Загружаем TLS credentials
	tlsConfig, err := utils.LoadTLSConfig()
	tlsCreds := credentials.NewTLS(tlsConfig)
	if err != nil {
		logger.Log.Error("failed to load TLS credentials", zap.Error(err))
		return "", -1, fmt.Errorf("Ошибка загрузки TLS credentials: %w", err), codes.Internal
	}

	// Создаем подключение с TLS
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(tlsCreds),
	)
	if err != nil {
		logger.Log.Error("grpc dial auth server failed", zap.Error(err), zap.String("addr", address))
		return "", -1, fmt.Errorf("Ошибка подключения к gRPC серверу: %w", err), codes.Internal
	}
	defer conn.Close()

	client := auth_pb.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	md := metadata.Pairs("authorization", accessToken)
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	// Вызов метода Validate без передачи параметров
	resp, err := client.Validate(ctx, &emptypb.Empty{})
	if err != nil {
		logger.Log.Warn("validate failed", zap.Error(err))
		return "", -1, err, codes.Unauthenticated
	}

	logger.Log.Debug("validate success", zap.String("role", resp.GetRole()), zap.Int32("user_id", resp.GetId()))
	return resp.GetRole(), int(resp.GetId()), nil, codes.OK
}

func RequestRefreshToken(refreshToken string) (string, error, codes.Code) {
	var address string = os.Getenv("AUTH_SERVER_GRPC_PORT")
	if address == "" {
		return "", fmt.Errorf("AUTH_SERVER_GRPC_PORT environment variable is not set"), codes.InvalidArgument
	}

	tlsConfig, err := utils.LoadTLSConfig()
	tlsCreds := credentials.NewTLS(tlsConfig)
	if err != nil {
		return "", fmt.Errorf("Ошибка загрузки TLS credentials: %w", err), codes.Internal
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(tlsCreds),
	)
	if err != nil {
		return "", fmt.Errorf("Ошибка подключения к gRPC серверу: %w", err), codes.Internal
	}
	defer conn.Close()

	client := auth_pb.NewAuthServiceClient(conn)

	// Создание метаданных с токеном
	md := metadata.New(map[string]string{"Authorization": refreshToken})

	// Добавление метаданных к контексту
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Вызов метода RefreshToken без передачи параметров
	resp, err := client.RefreshToken(ctx, &emptypb.Empty{})
	if err != nil {
		return "", err, codes.Unauthenticated
	}

	return resp.GetAccessToken(), nil, codes.OK
}
