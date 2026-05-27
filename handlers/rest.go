package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	trait_pb "github.com/Daniel3579/trait-service-sdk/gen"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ––––––––––––––––––––––––––––––––––––––––––

type UserTrait struct {
	UserId    int       `json:"user_id"`
	BirthDate time.Time `json:"birth_date"`
	Height    int       `json:"height"`
	Weight    int       `json:"weight"`
}

// ––––––––––––––––––––––––––––––––––––––––––

type HttpServer struct {
	GrpcSrv trait_pb.TraitServiceClient
}

// –––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––

func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// –––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––

func (h *HttpServer) Update(w http.ResponseWriter, r *http.Request) {
	var reqBody UserTrait
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	md := metadata.Pairs("authorization_access", accessToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcReq := &trait_pb.UserTrait{UserId: int32(reqBody.UserId), BirthDate: timestamppb.New(reqBody.BirthDate), Height: int32(reqBody.Height), Weight: int32(reqBody.Weight)}
	resp, err := h.GrpcSrv.Update(ctx, grpcReq)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true, // выводить нулевые поля
		UseProtoNames:   true, // snake_case как в proto
	}
	jsonBytes, err := marshaler.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
