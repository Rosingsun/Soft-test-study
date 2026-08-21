package service

import (
	"testing"

	"github.com/soft-test-study/backend/internal/model"
)

// TestIsAnswerCorrect 表驱动测试客观题判分函数
// 覆盖：单选 / 多选 / 判断 / 多空（JSON 数组）/ 错答 / 长度不一致 / 非法 JSON
func TestIsAnswerCorrect(t *testing.T) {
	tests := []struct {
		name      string
		typ       string
		correct   string
		user      string
		expected  bool
	}{
		{
			name:     "单选-正确",
			typ:      model.TypeSingle,
			correct:  "A",
			user:     "A",
			expected: true,
		},
		{
			name:     "单选-错误",
			typ:      model.TypeSingle,
			correct:  "A",
			user:     "B",
			expected: false,
		},
		{
			name:     "多选-正确（字符串相等）",
			typ:      model.TypeMulti,
			correct:  "AB",
			user:     "AB",
			expected: true,
		},
		{
			name:     "判断-正确",
			typ:      model.TypeJudge,
			correct:  "true",
			user:     "true",
			expected: true,
		},
		{
			name:     "判断-错误",
			typ:      model.TypeJudge,
			correct:  "true",
			user:     "false",
			expected: false,
		},
		{
			name:     "多空题-全部正确",
			typ:      model.TypeMultiBlank,
			correct:  `["A","B","C"]`,
			user:     `["A","B","C"]`,
			expected: true,
		},
		{
			name:     "多空题-部分错误",
			typ:      model.TypeMultiBlank,
			correct:  `["A","B","C"]`,
			user:     `["A","D","C"]`,
			expected: false,
		},
		{
			name:     "多空题-长度不一致",
			typ:      model.TypeMultiBlank,
			correct:  `["A","B","C"]`,
			user:     `["A","B"]`,
			expected: false,
		},
		{
			name:     "多空题-非法 JSON",
			typ:      model.TypeMultiBlank,
			correct:  `["A","B"]`,
			user:     `not a json`,
			expected: false,
		},
		{
			name:     "大小写敏感（单选）",
			typ:      model.TypeSingle,
			correct:  "A",
			user:     "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAnswerCorrect(tt.typ, tt.correct, tt.user)
			if got != tt.expected {
				t.Errorf("IsAnswerCorrect(%q, %q, %q) = %v, want %v",
					tt.typ, tt.correct, tt.user, got, tt.expected)
			}
		})
	}
}

// TestIsSubjectiveType 验证主观题识别（论文 / 案例分析不自动判分）
func TestIsSubjectiveType(t *testing.T) {
	tests := []struct {
		typ      string
		expected bool
	}{
		{model.TypeEssay, true},
		{model.TypeCaseStudy, true},
		{model.TypeSingle, false},
		{model.TypeMulti, false},
		{model.TypeJudge, false},
		{"unknown", false},
	}
	for _, tt := range tests {
		if got := IsSubjectiveType(tt.typ); got != tt.expected {
			t.Errorf("IsSubjectiveType(%q) = %v, want %v", tt.typ, got, tt.expected)
		}
	}
}
