package utils

import (
	"os"
	"testing"
	"trait-service/logger"
)

// ——————————————————————————————————————————————————————————————————————————————

func TestMain(m *testing.M) {
	// init logger
	if err := logger.Init(true); err != nil {
		panic(err)
	}
	code := m.Run()
	logger.Sync()
	os.Exit(code)
}

// ——————————————————————————————————————————————————————————————————————————————

func TestHashAndCheckPassword(t *testing.T) {

}

func TestGenerateAndValidateToken_Success(t *testing.T) {

}

func TestGenerateAndValidateToken_WrongType(t *testing.T) {

}

func TestGenerateAndValidateToken_Expired(t *testing.T) {

}

func TestGetTokenMetadata_FromContext(t *testing.T) {

}

func TestGetTokenMetadata_Missing(t *testing.T) {

}
