package handlers

import (
	"context"
	"trait-service/db"
	"trait-service/logger"

	trait_pb "github.com/Daniel3579/trait-service-sdk/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	trait_pb.UnimplementedTraitServiceServer
}

// Create – создание характеристик пользователя (только для админов и системы)
func (s *Server) Create(ctx context.Context, req *trait_pb.UserTrait) (*trait_pb.UserTrait, error) {
	role, ok := ctx.Value("role").(string)
	if !ok {
		logger.Log.Warn("role not found in context")
		return nil, status.Error(codes.Unauthenticated, "role not found in context")
	}

	if role == "user" {
		logger.Log.Warn("Unsuitable role for Create")
		return nil, status.Error(codes.PermissionDenied, "Access denied, unsuitable role")
	}

	res, err := db.Insert(req)
	if err != nil {
		logger.Log.Error("Insert into user_traits db error", zap.Error(err), zap.Int32("user_id", req.GetUserId()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Log.Info("UserTrait created", zap.Int32("user_id", res.GetUserId()))
	return res, nil
}

// Read – получение характеристик пользователя (только для админов и системы)
func (s *Server) Read(ctx context.Context, req *trait_pb.IdRequest) (*trait_pb.UserTrait, error) {
	// role, ok := ctx.Value("role").(string)
	// if !ok {
	// 	logger.Log.Warn("role not found in context")
	// 	return nil, status.Error(codes.Unauthenticated, "role not found in context")
	// }

	// if role == "user" {
	// 	logger.Log.Warn("Unsuitable role for Create")
	// 	return nil, status.Error(codes.PermissionDenied, "Access denied, unsuitable role")
	// }

	res, err := db.Select(req)
	if err != nil {
		logger.Log.Error("Select from user_traits db error", zap.Error(err), zap.Int32("user_id", req.GetUserId()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Log.Info("UserTrait read", zap.Int32("user_id", res.GetUserId()))
	return res, nil
}

// ReadMultiple – получение id пользователей по фильтрам характеристик (только для админов и системы)
func (s *Server) ReadMultiple(ctx context.Context, req *trait_pb.TraitRequest) (*trait_pb.MultipleReadResponse, error) {
	role, ok := ctx.Value("role").(string)
	if !ok {
		logger.Log.Warn("role not found in context")
		return nil, status.Error(codes.Unauthenticated, "role not found in context")
	}

	if role == "user" {
		logger.Log.Warn("Unsuitable role for Create")
		return nil, status.Error(codes.PermissionDenied, "Access denied, unsuitable role")
	}

	res, err := db.SelectMultiple(req)
	if err != nil {
		logger.Log.Error("SelectMultiple from user_traits db error", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Log.Info("Multiple UserTraits read", zap.Int("count", len(res.GetUserIds())))
	return res, nil
}

// Update – обновление характеристик пользователя
// - админ может обновить любые
// - обычный пользователь может обновить только свои
func (s *Server) Update(ctx context.Context, req *trait_pb.UserTrait) (*trait_pb.UserTrait, error) {
	userIDFromCtx, ok := ctx.Value("user_id").(int)
	if !ok {
		logger.Log.Warn("user_id not found in context")
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	role, ok := ctx.Value("role").(string)
	if !ok {
		logger.Log.Warn("role not found in context")
		return nil, status.Error(codes.Unauthenticated, "role not found in context")
	}

	if role == "user" {
		if userIDFromCtx != int(req.GetUserId()) {
			logger.Log.Warn("Token user_id mismatch with request user_id",
				zap.Int("token_user_id", userIDFromCtx),
				zap.Int32("req_user_id", req.GetUserId()))
			return nil, status.Error(codes.PermissionDenied, "Access denied, user_id mismatch")
		}
	}

	res, err := db.Update(req)
	if err != nil {
		logger.Log.Error("Update user_traits db error", zap.Error(err), zap.Int32("user_id", req.GetUserId()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Log.Info("UserTrait updated", zap.Int32("user_id", res.GetUserId()))
	return res, nil
}

// Delete – удаление характеристик пользователя (только для админов и системы)
func (s *Server) Delete(ctx context.Context, req *trait_pb.IdRequest) (*trait_pb.UserTrait, error) {
	role, ok := ctx.Value("role").(string)
	if !ok {
		logger.Log.Warn("role not found in context")
		return nil, status.Error(codes.Unauthenticated, "role not found in context")
	}

	if role == "user" {
		logger.Log.Warn("Unsuitable role for Delete")
		return nil, status.Error(codes.PermissionDenied, "Access denied, unsuitable role")
	}

	res, err := db.Delete(req)
	if err != nil {
		logger.Log.Error("Delete user_traits db error", zap.Error(err), zap.Int32("user_id", req.GetUserId()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Log.Info("UserTrait deleted", zap.Int32("user_id", res.GetUserId()))
	return res, nil
}
