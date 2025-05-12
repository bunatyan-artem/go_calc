package helpers

import (
	"calculator/internal/orchestrator/Tree"
	"github.com/dgrijalva/jwt-go"
	"testing"
)

func TestGenerateJWT(t *testing.T) {
	var testUsers = []string{"testuser", "cfgvbhnj", "664830g", "_log"}

	for _, testUser := range testUsers {
		tokenString, err := GenerateJWT(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("very_big_secret"), nil
		})
		if err != nil || !token.Valid {
			t.Fatalf("Token is invalid: %v", err)
		}

		claims := token.Claims.(jwt.MapClaims)
		if claims["login"] != testUser {
			t.Fatalf("Expected login 'testuser', got %v", claims["login"])
		}
	}
}

func TestExpressionInfo(t *testing.T) {
	Trees = []*Tree.Tree{
		{Flag: 1},
		{Flag: 2, Result: 42.0},
		{Flag: 3},
	}

	tests := []struct {
		id     int
		status string
		result float64
	}{
		{0, "processing", 0},
		{1, "done", 42.0},
		{2, "error: division by zero", 0},
	}

	for _, tt := range tests {
		info := ExpressionInfo(tt.id)
		if info.Status != tt.status || info.Result != tt.result {
			t.Errorf("expressionInfo(%d): got %+v", tt.id, info)
		}
	}
}
