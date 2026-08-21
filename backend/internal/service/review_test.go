package service

import (
	"testing"
	"time"

	"github.com/soft-test-study/backend/internal/model"
)

// newTestCard 构造一张默认参数的 ReviewCard 用于测试
func newTestCard() *model.ReviewCard {
	return &model.ReviewCard{
		EaseFactor:   2.5,
		Repetition:   0,
		IntervalDays: 0,
		DueDate:      time.Now(),
	}
}

// TestApplySM2_AnswerCorrect quality>=3 时按 SM-2 标准间隔递增
// 验证：
//   - 首次答对 repetition 0 -> 1，IntervalDays 固定为 1
//   - 第二次答对 repetition 1 -> 2，IntervalDays 固定为 6
//   - 第三次答对按 EF 倍率递增
func TestApplySM2_AnswerCorrect(t *testing.T) {
	card := newTestCard()

	// 第一次答对
	applySM2(card, 5)
	if card.Repetition != 1 {
		t.Errorf("首次答对 Repetition 期望 1，实际 %d", card.Repetition)
	}
	if card.IntervalDays != 1 {
		t.Errorf("首次答对 IntervalDays 期望 1，实际 %d", card.IntervalDays)
	}
	firstEF := card.EaseFactor
	if firstEF < reviewEfMin || firstEF > reviewEfMax {
		t.Errorf("EF 越界: %f", firstEF)
	}

	// 第二次答对
	applySM2(card, 5)
	if card.Repetition != 2 {
		t.Errorf("第二次答对 Repetition 期望 2，实际 %d", card.Repetition)
	}
	if card.IntervalDays != 6 {
		t.Errorf("第二次答对 IntervalDays 期望 6，实际 %d", card.IntervalDays)
	}

	// 第三次答对：按 EF 递增（applySM2 中 EF 在每次调用时 +0.1，第三次 local ef = 2.7+0.1=2.8）
	applySM2(card, 5)
	if card.Repetition != 3 {
		t.Errorf("第三次答对 Repetition 期望 3，实际 %d", card.Repetition)
	}
	expected := 17 // round(6 * 2.8) = 17，applySM2 在第三次调用时 local ef = 2.7 + 0.1 = 2.8
	if expected > reviewMaxInterval {
		expected = reviewMaxInterval
	}
	if card.IntervalDays != expected {
		t.Errorf("第三次答对 IntervalDays 期望 %d，实际 %d", expected, card.IntervalDays)
	}
}

// TestApplySM2_AnswerWrong quality<3 时重置 repetition 和 interval
func TestApplySM2_AnswerWrong(t *testing.T) {
	card := newTestCard()
	card.Repetition = 5
	card.IntervalDays = 90

	applySM2(card, 2)

	if card.Repetition != 0 {
		t.Errorf("答错 Repetition 期望 0，实际 %d", card.Repetition)
	}
	if card.IntervalDays != 1 {
		t.Errorf("答错 IntervalDays 期望 1，实际 %d", card.IntervalDays)
	}
}

// TestApplySM2_EFBounds 验证 EF 下限：连续答错 EF 不会低于 reviewEfMin
func TestApplySM2_EFBounds(t *testing.T) {
	card := newTestCard()
	card.EaseFactor = 1.3

	// 答错多次（quality=0），EF 应被钳制到 reviewEfMin
	for i := 0; i < 5; i++ {
		applySM2(card, 0)
	}
	if card.EaseFactor != reviewEfMin {
		t.Errorf("EF 下限期望 %f，实际 %f", reviewEfMin, card.EaseFactor)
	}
}

// TestApplySM2_IntervalCap 验证 IntervalDays 上限
func TestApplySM2_IntervalCap(t *testing.T) {
	card := newTestCard()
	card.Repetition = 10
	card.IntervalDays = reviewMaxInterval * 2
	card.EaseFactor = 3.0

	applySM2(card, 5)

	if card.IntervalDays != reviewMaxInterval {
		t.Errorf("IntervalDays 上限期望 %d，实际 %d", reviewMaxInterval, card.IntervalDays)
	}
}

// TestApplySM2_Quality3_Boundary 验证 quality=3 仍算答对（>=3）
func TestApplySM2_Quality3_Boundary(t *testing.T) {
	card := newTestCard()
	card.Repetition = 1
	card.IntervalDays = 6

	applySM2(card, 3)

	if card.Repetition != 2 {
		t.Errorf("quality=3 视为答对，Repetition 期望 2，实际 %d", card.Repetition)
	}
}
