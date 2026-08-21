package security

import (
	"errors"
	"strings"
	"testing"
	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword_ValidPassword 正常密码能成功哈希
func TestHashPassword_ValidPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("哈希结果为空")
	}
	// bcrypt 默认 cost=10
	if !strings.HasPrefix(hash, "$2a$10$") && !strings.HasPrefix(hash, "$2b$10$") {
		t.Errorf("哈希应以 $2a$10$ 开头（bcrypt default cost），实际: %s", hash[:10])
	}
}

// TestVerifyPassword_Correct 正确密码校验通过
func TestVerifyPassword_Correct(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if err := VerifyPassword(hash, "correct-password"); err != nil {
		t.Errorf("正确密码应校验通过，但返回: %v", err)
	}
}

// TestVerifyPassword_Wrong 错误密码校验失败
func TestVerifyPassword_Wrong(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if err := VerifyPassword(hash, "wrong-password"); err == nil {
		t.Error("错误密码应校验失败，但返回 nil")
	} else if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Errorf("错误密码应返回 bcrypt.ErrMismatchedHashAndPassword，实际: %v", err)
	}
}

// TestVerifyPassword_EmptyHash 非法 hash 返回错误
func TestVerifyPassword_EmptyHash(t *testing.T) {
	if err := VerifyPassword("", "any-password"); err == nil {
		t.Error("空 hash 应返回错误，但返回 nil")
	}
}

// TestHashPassword_DifferentResults 同一密码两次哈希结果不同（bcrypt salt）
func TestHashPassword_DifferentResults(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	if h1 == h2 {
		t.Error("同一密码两次哈希应不同（bcrypt 加盐），但实际相同")
	}
	// 但两者都应该能校验通过
	if err := VerifyPassword(h1, "same-password"); err != nil {
		t.Errorf("第一次哈希校验失败: %v", err)
	}
	if err := VerifyPassword(h2, "same-password"); err != nil {
		t.Errorf("第二次哈希校验失败: %v", err)
	}
}
