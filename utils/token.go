package utils

import (
	"context"
	"fmt"
	"trait-service/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func GetTokenMetadata(ctx context.Context, metaType string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Log.Warn("missing metadata in context")
		return "", fmt.Errorf("Missing metadata: %v", ok)
	}

	var tokens []string = md[metaType]
	if len(tokens) == 0 {
		logger.Log.Warn("missing token", zap.String("type", metaType))
		return "", fmt.Errorf("Missing %s token", metaType)
	}

	return tokens[0], nil
}
