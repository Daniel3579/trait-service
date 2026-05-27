package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"
	"trait-service/logger"

	trait_pb "github.com/Daniel3579/trait-service-sdk/gen"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var db *sql.DB

// ——————————————————————————————————————————————————————————————————————————————

func ConnectDB(env string) error {
	var connStr string = os.Getenv(env)

	if connStr == "" {
		logger.Log.Error("DATABASE_URL environment variable is not set")
		return fmt.Errorf("DATABASE_URL not set")
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Log.Error("Failed to open database", zap.Error(err))
		return fmt.Errorf("Ошибка при открытии базы данных: %w", err)
	}

	err = db.Ping()
	if err != nil {
		logger.Log.Error("Failed to ping database", zap.Error(err))
		return fmt.Errorf("Не удалось пингануть бд: %w", err)
	}

	logger.Log.Info("Successfully connected to database")
	return nil
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// ——————————————————————————————————————————————————————————————————————————————

func Insert(req *trait_pb.UserTrait) (*trait_pb.UserTrait, error) {
	err := db.QueryRow(`
		INSERT INTO user_traits (user_id, birth_date, height, weight)
		VALUES ($1, $2, $3, $4)
		RETURNING *;`,
		req.GetUserId(), nil, 0, 0,
	).Scan(&req.UserId, &req.BirthDate, &req.Height, &req.Weight)

	if err != nil {
		logger.Log.Error("Failed to insert into user_traits",
			zap.Int32("user_id", req.GetUserId()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("не удалось записать в БД: %w", err)
	}

	logger.Log.Info("UserTrait created successfully", zap.Int32("user_id", req.GetUserId()))
	return req, nil
}

func Select(req *trait_pb.IdRequest) (*trait_pb.UserTrait, error) {
	var res trait_pb.UserTrait
	var birth sql.NullTime

	err := db.QueryRow(`
		SELECT *
		FROM user_traits
		WHERE user_id = $1;`,
		req.GetUserId(),
	).Scan(&res.UserId, &birth, &res.Height, &res.Weight)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn("UserTrait not found", zap.Int32("user_id", req.GetUserId()))
			return nil, fmt.Errorf("UserTrait not found")
		}
		logger.Log.Error("Failed to query UserTrait",
			zap.Int32("user_id", req.GetUserId()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("не удалось получить UserTrait: %w", err)
	}

	if birth.Valid {
		res.BirthDate = timestamppb.New(birth.Time)
	} else {
		res.BirthDate = nil
	}

	logger.Log.Info("UserTrait selected successfully")
	return &res, nil
}

func SelectMultiple(req *trait_pb.TraitRequest) (*trait_pb.MultipleReadResponse, error) {
	query := `
		SELECT user_id
		FROM user_traits
		WHERE 1=1`
	args := []interface{}{}
	argCounter := 1

	// Фильтр по дате рождения
	if req.BirthFrom != nil {
		query += fmt.Sprintf(" AND birth_date <= $%d", argCounter)
		args = append(args, req.BirthTo.AsTime())
		argCounter++
	}
	if req.BirthTo != nil {
		query += fmt.Sprintf(" AND birth_date <= $%d", argCounter)
		args = append(args, req.BirthTo.AsTime())
		argCounter++
	}

	// Фильтр по росту
	if req.HeightFrom != 0 {
		query += fmt.Sprintf(" AND height >= $%d", argCounter)
		args = append(args, req.HeightFrom)
		argCounter++
	}
	if req.HeightTo != 0 {
		query += fmt.Sprintf(" AND height <= $%d", argCounter)
		args = append(args, req.HeightTo)
		argCounter++
	}

	// Фильтр по весу
	if req.WeightFrom != 0 {
		query += fmt.Sprintf(" AND weight >= $%d", argCounter)
		args = append(args, req.WeightFrom)
		argCounter++
	}
	if req.WeightTo != 0 {
		query += fmt.Sprintf(" AND weight <= $%d", argCounter)
		args = append(args, req.WeightTo)
		argCounter++
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Log.Error("Failed to query user_traits for SelectMultiple", zap.Error(err))
		return nil, fmt.Errorf("не удалось выполнить поиск: %w", err)
	}
	defer rows.Close()

	var userIDs []*trait_pb.SingleRead
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			logger.Log.Error("Failed to scan row in SelectMultiple", zap.Error(err))
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		userIDs = append(userIDs, &trait_pb.SingleRead{UserId: id})
	}
	if err = rows.Err(); err != nil {
		logger.Log.Error("Rows iteration error in SelectMultiple", zap.Error(err))
		return nil, fmt.Errorf("ошибка при обработке результатов: %w", err)
	}

	logger.Log.Info("SelectMultiple completed", zap.Int("count", len(userIDs)))
	return &trait_pb.MultipleReadResponse{UserIds: userIDs}, nil
}

func Update(req *trait_pb.UserTrait) (*trait_pb.UserTrait, error) {
	var birthArg interface{}
	if req.BirthDate != nil {
		birthArg = req.BirthDate.AsTime()
	} else {
		birthArg = nil
	}

	var birth sql.NullTime

	err := db.QueryRow(`
		UPDATE user_traits
		SET birth_date = $1, height = $2, weight = $3
		WHERE user_id = $4
		RETURNING *;`,
		birthArg, req.GetHeight(), req.GetWeight(), req.GetUserId(),
	).Scan(&req.UserId, &birth, &req.Height, &req.Weight)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn("UserTrait not found for update", zap.Int32("user_id", req.GetUserId()))
			return nil, fmt.Errorf("UserTrait not found")
		}
		logger.Log.Error("Failed to update UserTrait",
			zap.Int32("user_id", req.GetUserId()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("не удалось обновить запись: %w", err)
	}

	if birth.Valid {
		req.BirthDate = timestamppb.New(birth.Time)
	} else {
		req.BirthDate = nil
	}

	logger.Log.Info("UserTrait updated successfully", zap.Int32("user_id", req.GetUserId()))
	return req, nil
}

func Delete(req *trait_pb.IdRequest) (*trait_pb.UserTrait, error) {
	var res trait_pb.UserTrait
	var birth time.Time

	err := db.QueryRow(`
		DELETE FROM user_traits
		WHERE user_id = $1
		RETURNING *;`,
		req.GetUserId(),
	).Scan(&res.UserId, &birth, &res.Height, &res.Weight)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn("UserTrait not found for delete", zap.Int32("user_id", req.GetUserId()))
			return nil, fmt.Errorf("UserTrait not found")
		}
		logger.Log.Error("Failed to delete UserTrait",
			zap.Int32("user_id", req.GetUserId()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("не удалось удалить запись: %w", err)
	}

	if !birth.IsZero() {
		res.BirthDate = timestamppb.New(birth)
	}
	logger.Log.Info("UserTrait deleted successfully", zap.Int32("user_id", req.GetUserId()))
	return &res, nil
}
