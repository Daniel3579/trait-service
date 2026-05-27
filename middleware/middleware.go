package middleware

import (
	"context"
	"time"
	"trait-service/logger"

	auth "trait-service/auth"
	"trait-service/utils"

	"google.golang.org/grpc/codes"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryLoggingInterceptor перехватывает все unary RPC запросы и логирует их
func UnaryMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := "OK"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code().String()
			} else {
				code = "Unknown"
			}
			logger.Log.Warn("rpc error",
				zap.String("method", info.FullMethod),
				zap.String("code", code),
				zap.Error(err),
			)
		} else {
			logger.Log.Info("rpc completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", time.Since(start)),
			)
		}

		return resp, err
	}
}

func ValidateMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	logger.Log.Debug("incoming request", zap.String("method", info.FullMethod))

	token, err := utils.GetTokenMetadata(ctx, "authorization_access")
	if err != nil {
		logger.Log.Warn("missing access token", zap.Error(err), zap.String("method", info.FullMethod))
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	role, user_id, err, errCode := auth.RequestValidate(token)
	if err != nil {
		if errCode == codes.Unauthenticated {
			logger.Log.Info("access token invalid, attempting refresh", zap.String("method", info.FullMethod))
			refreshToken, err := utils.GetTokenMetadata(ctx, "authorization_refresh")
			if err != nil {
				logger.Log.Warn("missing refresh token", zap.Error(err))
				return nil, status.Error(codes.Unauthenticated, err.Error())
			}

			accessToken, err, errCode := auth.RequestRefreshToken(refreshToken)
			if err != nil {
				logger.Log.Error("refresh token failed", zap.Error(err))
				return nil, status.Error(errCode, err.Error())
			}

			role, user_id, err, errCode = auth.RequestValidate(accessToken)
			if err != nil {
				logger.Log.Error("validate after refresh failed", zap.Error(err))
				return nil, status.Error(errCode, err.Error())
			}

		} else {
			logger.Log.Error("validate token error", zap.Error(err))
			return nil, status.Error(errCode, err.Error())
		}
	}

	// Установите имя пользователя в контекст для использования в следующем обработчике
	ctx = context.WithValue(ctx, "role", role)
	ctx = context.WithValue(ctx, "user_id", user_id)
	logger.Log.Debug("authenticated request", zap.String("role", role), zap.Int("user_id", user_id), zap.String("method", info.FullMethod))

	// Вызовите следующий обработчик в цепочке
	return handler(ctx, req)
}
