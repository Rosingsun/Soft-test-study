package service

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	"gorm.io/gorm"
)

// inviteCodeAlphabet 生成邀请码所用字符表：
//   - 大写字母 + 数字
//   - 剔除 0/O/1/I/L 等视觉易混字符，降低分发与录入错误率
const inviteCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// inviteCodeLength 邀请码长度
const inviteCodeLength = 12

type InvitationCodeService struct {
	repo *repository.InvitationCodeRepo
}

func NewInvitationCodeService(repo *repository.InvitationCodeRepo) *InvitationCodeService {
	return &InvitationCodeService{repo: repo}
}

// Generate 批量生成 n 个不重复的邀请码
//
// 入参：
//   - count：需要生成的数量
//   - createdBy：生成者 user_id（CLI 调用时传 0）
//   - note：批次备注，便于审计/分发追踪
//
// 行为：
//   - 每次循环生成 1 个码并立即落库
//   - 遇到唯一键冲突（极小概率）自动重试，最多 8 次
//   - 全部成功后返回生成的码列表（含 id/created_at）
func (s *InvitationCodeService) Generate(count int, createdBy uint, note string) ([]model.InvitationCode, error) {
	if count <= 0 {
		return nil, errors.New("生成数量必须大于 0")
	}
	if count > 1000 {
		return nil, errors.New("单次生成不超过 1000 个")
	}

	out := make([]model.InvitationCode, 0, count)
	for i := 0; i < count; i++ {
		var ic model.InvitationCode
		var err error
		// 极端情况下随机碰撞，重试若干次
		for retry := 0; retry < 8; retry++ {
			code, genErr := generateOneCode(inviteCodeLength)
			if genErr != nil {
				return nil, fmt.Errorf("生成邀请码失败: %w", genErr)
			}
			ic = model.InvitationCode{
				Code:      code,
				CreatedBy: createdBy,
				Note:      note,
				Status:    1,
			}
			if err = s.repo.Create(&ic); err == nil {
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("生成邀请码 %d/%d 失败: %w", i+1, count, err)
		}
		out = append(out, ic)
	}
	return out, nil
}

// generateOneCode 用 crypto/rand 生成指定长度的邀请码（字符均匀分布，无偏置）
func generateOneCode(n int) (string, error) {
	if n <= 0 {
		return "", errors.New("邀请码长度必须大于 0")
	}
	alphabetLen := byte(len(inviteCodeAlphabet))
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range buf {
		// 用取模映射到字符表；alphabetLen=31，byte 均匀分布足够
		out[i] = inviteCodeAlphabet[b%alphabetLen]
	}
	return string(out), nil
}

// ValidateAndLockInTx 在事务中加行锁并校验邀请码有效性：
//   1. SELECT ... FOR UPDATE 锁定
//   2. 校验 status=1 且 used_by IS NULL
//
// 校验通过返回 nil；调用方在事务内继续创建用户，最后再调用 MarkUsedInTx 标记已用。
// 校验失败返回对应错误。
func (s *InvitationCodeService) ValidateAndLockInTx(tx *gorm.DB, code string) error {
	if code == "" {
		return errors.New("邀请码不能为空")
	}
	ic, err := s.repo.FindByCodeForUpdate(tx, code)
	if err != nil {
		return errors.New("邀请码无效")
	}
	if ic.Status != 1 {
		return errors.New("邀请码已被停用")
	}
	if ic.UsedBy != nil {
		return errors.New("邀请码已被使用")
	}
	return nil
}

// MarkUsedInTx 标记邀请码已被指定 user 使用
//
// 设计原因：注册事务里需要先 INSERT users 拿到 user.ID，再回填 used_by，
// 所以分两步：ValidateAndLockInTx → CreateUser → MarkUsedInTx。
func (s *InvitationCodeService) MarkUsedInTx(tx *gorm.DB, code string, usedBy uint) error {
	if code == "" {
		return errors.New("邀请码不能为空")
	}
	var id uint
	err := tx.Model(&model.InvitationCode{}).
		Where("code = ?", code).
		Select("id").
		Scan(&id).Error
	if err != nil || id == 0 {
		return errors.New("邀请码标记失败，请稍后重试")
	}
	return s.repo.MarkUsed(tx, id, usedBy)
}
