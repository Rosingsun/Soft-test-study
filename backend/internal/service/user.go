package service

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/model"
	"github.com/soft-test-study/backend/internal/repository"
	jwtutil "github.com/soft-test-study/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo        *repository.UserRepo
	levelRepo   *repository.ExamLevelRepo
	subjectRepo *repository.SubjectRepo
	jwtSecret   string
	jwtExpires  int
}

func NewUserService(repo *repository.UserRepo, levelRepo *repository.ExamLevelRepo, subjectRepo *repository.SubjectRepo, jwtSecret string, jwtExpires int) *UserService {
	return &UserService{repo: repo, levelRepo: levelRepo, subjectRepo: subjectRepo, jwtSecret: jwtSecret, jwtExpires: jwtExpires}
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度至少8位")
	}
	if len(password) > 128 {
		return errors.New("密码不超过128个字符")
	}
	var hasDigit bool
	for _, ch := range password {
		if unicode.IsDigit(ch) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}
	return nil
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return errors.New("用户名长度须在3-50个字符之间")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\p{Han}]+$`, username)
	if !matched {
		return errors.New("用户名只能包含字母、数字、下划线和中文")
	}
	return nil
}

func (s *UserService) validateSubjectLevel(levelID, subjectID uint) error {
	if levelID == 0 {
		return errors.New("请选择考试等级")
	}
	if subjectID == 0 {
		return errors.New("请选择考试科目")
	}
	return s.validateSubjectBelongsToLevel(levelID, subjectID)
}

func (s *UserService) validateSubjectBelongsToLevel(levelID, subjectID uint) error {
	if subjectID == 0 {
		return nil
	}
	if levelID == 0 {
		return errors.New("请选择考试等级")
	}
	if _, err := s.levelRepo.FindByID(levelID); err != nil {
		return errors.New("考试等级不存在")
	}
	subject, err := s.subjectRepo.FindByID(subjectID)
	if err != nil {
		return errors.New("考试科目不存在")
	}
	if subject.LevelID != levelID {
		return errors.New("所选科目不属于该考试等级")
	}
	return nil
}

func (s *UserService) buildUserInfo(user *model.User) *dto.UserInfo {
	info := &dto.UserInfo{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Nickname:   user.Nickname,
		Avatar:     user.Avatar,
		Role:       user.Role,
		LevelID:    user.LevelID,
		SubjectID:  user.SubjectID,
		Difficulty: user.Difficulty,
	}
	if user.LevelID > 0 {
		if level, err := s.levelRepo.FindByID(user.LevelID); err == nil {
			info.LevelName = level.Name
		}
	}
	if user.SubjectID > 0 {
		if subject, err := s.subjectRepo.FindByID(user.SubjectID); err == nil {
			info.SubjectName = subject.Name
		}
	}
	return info
}

func (s *UserService) Register(req *dto.RegisterReq) (*dto.LoginResp, error) {
	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("两次密码输入不一致")
	}
	if err := s.validateSubjectLevel(req.LevelID, req.SubjectID); err != nil {
		return nil, err
	}

	existing, _ := s.repo.FindByUsername(req.Username)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("用户名已被注册")
	}

	existing, _ = s.repo.FindByEmail(req.Email)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("邮箱已被注册")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         "student",
		Status:       1,
		LevelID:      req.LevelID,
		SubjectID:    req.SubjectID,
		Difficulty:   req.Difficulty,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("注册失败，请稍后重试")
	}

	expiresIn := time.Duration(s.jwtExpires) * time.Hour
	tokenStr, err := jwtutil.Generate(s.jwtSecret, user.ID, expiresIn)
	if err != nil {
		return nil, errors.New("Token 签发失败")
	}

	return &dto.LoginResp{
		AccessToken: tokenStr,
		ExpiresIn:   s.jwtExpires * 3600,
		UserInfo:    *s.buildUserInfo(user),
	}, nil
}

func (s *UserService) Login(req *dto.LoginReq) (*dto.LoginResp, error) {
	user, err := s.repo.FindByUsernameOrEmail(req.Username)
	if err != nil {
		return nil, errors.New("用户名/邮箱或密码错误")
	}

	if user.Status == 0 {
		return nil, errors.New("账号已被禁用")
	}
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remaining := time.Until(*user.LockedUntil).Round(time.Second)
		return nil, fmt.Errorf("账号已被锁定，请在 %s 后重试", remaining.String())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		user.FailedAttempts++
		if user.FailedAttempts >= 5 {
			now := time.Now()
			lockedUntil := now.Add(15 * time.Minute)
			user.LockedUntil = &lockedUntil
			user.FailedAttempts = 0
			s.repo.Update(user)
			return nil, errors.New("密码错误次数过多，账号已被锁定15分钟")
		}
		s.repo.Update(user)
		return nil, errors.New("用户名/邮箱或密码错误")
	}

	user.FailedAttempts = 0
	user.LockedUntil = nil
	s.repo.Update(user)

	expiresIn := time.Duration(s.jwtExpires) * time.Hour
	tokenStr, err := jwtutil.Generate(s.jwtSecret, user.ID, expiresIn)
	if err != nil {
		return nil, errors.New("Token 签发失败")
	}

	return &dto.LoginResp{
		AccessToken: tokenStr,
		ExpiresIn:   s.jwtExpires * 3600,
		UserInfo:    *s.buildUserInfo(user),
	}, nil
}

func (s *UserService) GetUserInfo(userID uint) (*dto.UserInfo, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return s.buildUserInfo(user), nil
}

func (s *UserService) UpdateProfile(userID uint, req *dto.UpdateProfileReq) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.LevelID != nil {
		user.LevelID = *req.LevelID
	}
	if req.SubjectID != nil {
		user.SubjectID = *req.SubjectID
	}
	if req.Difficulty != nil {
		user.Difficulty = *req.Difficulty
	}
	if req.LevelID != nil || req.SubjectID != nil {
		if err := s.validateSubjectBelongsToLevel(user.LevelID, user.SubjectID); err != nil {
			return err
		}
	}
	return s.repo.Update(user)
}

func (s *UserService) ChangePassword(userID uint, req *dto.ChangePasswordReq) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	if err := validatePassword(req.NewPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user.PasswordHash = string(hash)
	return s.repo.Update(user)
}
