package service

import "testing"

// TestValidatePassword 表驱动测试密码复杂度校验
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		expectErr bool
	}{
		{"太短", "abc1", true},
		{"仅8位无数字", "abcdefgh", true},
		{"8位含数字", "abcdefg1", false},
		{"含中文与数字", "密码123", false},
		{"正常密码", "Password123", false},
		{"128位边界", "a1" + repeat(126, "a"), false},
		{"超长129位", "a1" + repeat(127, "a"), true},
		{"空字符串", "", true},
		{"7位+数字", "abcd1ab", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.expectErr && err == nil {
				t.Errorf("validatePassword(%q) 期望返回错误，但返回 nil", tt.password)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("validatePassword(%q) 期望通过，但返回错误: %v", tt.password, err)
			}
		})
	}
}

// repeat 简单字符串重复
func repeat(n int, s string) string {
	out := make([]byte, 0, n*len(s))
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
