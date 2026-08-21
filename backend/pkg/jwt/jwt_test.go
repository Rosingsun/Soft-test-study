package jwt

import (
	"testing"
	"time"
)

// TestGenerateAndParse 正常路径：生成 token 后能解析出原 userID
func TestGenerateAndParse(t *testing.T) {
	secret := "test-secret-1234567890"
	userID := uint(42)
	token, err := Generate(secret, userID, time.Hour)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if token == "" {
		t.Fatal("Generate 返回空 token")
	}

	got, err := Parse(secret, token)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if got != userID {
		t.Errorf("Parse userID 期望 %d，实际 %d", userID, got)
	}
}

// TestParse_WrongSecret 错误密钥应返回错误
func TestParse_WrongSecret(t *testing.T) {
	token, err := Generate("secret-a-1234567890", 1, time.Hour)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := Parse("secret-b-1234567890", token); err == nil {
		t.Error("错误密钥应返回错误，但解析成功")
	}
}

// TestParse_Expired 过期 token 应返回错误
func TestParse_Expired(t *testing.T) {
	secret := "test-secret-1234567890"
	token, err := Generate(secret, 1, -time.Second) // 立即过期
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := Parse(secret, token); err == nil {
		t.Error("过期 token 应返回错误，但解析成功")
	}
}

// TestParse_MalformedToken 格式错误的 token 应返回错误
func TestParse_MalformedToken(t *testing.T) {
	if _, err := Parse("any-secret", "not-a-jwt"); err == nil {
		t.Error("非法 token 应返回错误，但解析成功")
	}
	if _, err := Parse("any-secret", ""); err == nil {
		t.Error("空 token 应返回错误，但解析成功")
	}
}

// TestGenerate_DifferentUserIDs 不同 userID 生成的 token 不一致
func TestGenerate_DifferentUserIDs(t *testing.T) {
	secret := "test-secret-1234567890"
	t1, _ := Generate(secret, 1, time.Hour)
	t2, _ := Generate(secret, 2, time.Hour)
	if t1 == t2 {
		t.Error("不同 userID 生成的 token 应不同")
	}
}
