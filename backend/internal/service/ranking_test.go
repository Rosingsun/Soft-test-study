package service

import (
	"math/rand"
	"sort"
	"testing"
)

// makeBenchItems 生成 n 个随机 rankValue，按 (SortVal DESC, SubVal DESC) 排序
func makeBenchItems(n int) []rankValue {
	items := make([]rankValue, n)
	for i := 0; i < n; i++ {
		items[i] = rankValue{
			UserID:  uint(i + 1),
			Value:   rand.Float64() * 100,
			SortVal: rand.Float64() * 100,
			SubVal:  rand.Float64(),
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].SortVal != items[j].SortVal {
			return items[i].SortVal > items[j].SortVal
		}
		return items[i].SubVal > items[j].SubVal
	})
	return items
}

// findMyRankLinear 旧的线性实现（仅用于 benchmark 对比，不在生产路径使用）
func findMyRankLinear(items []rankValue, userID uint) (rank int, myValue float64, found bool) {
	var my *rankValue
	for i := range items {
		if items[i].UserID == userID {
			my = &items[i]
			break
		}
	}
	if my == nil {
		return 0, 0, false
	}
	r := 1
	for _, it := range items {
		if it.SortVal > my.SortVal {
			r++
		} else if it.SortVal == my.SortVal && it.SubVal > my.SubVal {
			r++
		} else {
			break
		}
	}
	return r, my.Value, true
}

// TestFindMyRank 表驱动测试 findMyRank 的正确性（OPT-24 验收）
func TestFindMyRank(t *testing.T) {
	items := []rankValue{
		{UserID: 1, Value: 90, SortVal: 90, SubVal: 5},
		{UserID: 2, Value: 80, SortVal: 80, SubVal: 8},
		{UserID: 3, Value: 80, SortVal: 80, SubVal: 6}, // 与 user 2 同分但 SubVal 较小
		{UserID: 4, Value: 70, SortVal: 70, SubVal: 7},
		{UserID: 5, Value: 60, SortVal: 60, SubVal: 9},
	}
	// 已按 (SortVal DESC, SubVal DESC) 排好

	tests := []struct {
		name     string
		userID   uint
		wantRank int
		wantVal  float64
		wantOK   bool
	}{
		{"第一名", 1, 1, 90, true},
		{"第二名（SubVal 较高）", 2, 2, 80, true},
		{"第三名（与 user 2 同 SortVal、SubVal 较小）", 3, 3, 80, true}, // 旧实现因 SubVal 次序排在第 3
		{"中间名次", 4, 4, 70, true},
		{"末位", 5, 5, 60, true},
		{"不存在", 999, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, val, ok := findMyRank(items, tt.userID)
			if ok != tt.wantOK {
				t.Errorf("found 期望 %v，实际 %v", tt.wantOK, ok)
			}
			if rank != tt.wantRank {
				t.Errorf("rank 期望 %d，实际 %d", tt.wantRank, rank)
			}
			if val != tt.wantVal {
				t.Errorf("value 期望 %v，实际 %v", tt.wantVal, val)
			}
		})
	}
}

// TestFindMyRank_NewEqualsOld 对随机数据验证新实现与旧实现结果一致
func TestFindMyRank_NewEqualsOld(t *testing.T) {
	rand.Seed(42) // 旧 API，Go 1.20+ 仍兼容
	for _, n := range []int{1, 10, 100, 1000} {
		items := makeBenchItems(n)
		// 选一个中间位置
		target := uint(n / 2)
		rNew, vNew, fNew := findMyRank(items, target)
		rOld, vOld, fOld := findMyRankLinear(items, target)
		if fNew != fOld || rNew != rOld || vNew != vOld {
			t.Errorf("n=%d target=%d: new=(%d,%v,%v) old=(%d,%v,%v)",
				n, target, rNew, vNew, fNew, rOld, vOld, fOld)
		}
	}
}

// BenchmarkFindMyRank 新实现（sort.Search + slices.IndexFunc），10000 用户
func BenchmarkFindMyRank(b *testing.B) {
	items := makeBenchItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = findMyRank(items, 5000)
	}
}

// BenchmarkFindMyRankLinear 旧实现（线性扫描），10000 用户
func BenchmarkFindMyRankLinear(b *testing.B) {
	items := makeBenchItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = findMyRankLinear(items, 5000)
	}
}

// BenchmarkFindMyRank_Last 末位查找（最坏情况：my 在排序末尾）
func BenchmarkFindMyRank_Last(b *testing.B) {
	items := makeBenchItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = findMyRank(items, 10000)
	}
}

func BenchmarkFindMyRankLinear_Last(b *testing.B) {
	items := makeBenchItems(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = findMyRankLinear(items, 10000)
	}
}
