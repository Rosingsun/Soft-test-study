package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/soft-test-study/backend/internal/dto"
	"github.com/soft-test-study/backend/internal/service"
	"github.com/soft-test-study/backend/pkg/response"
)

type UserHandler struct {
	svc        *service.UserService
	emailSvc   *service.EmailService
}

func NewUserHandler(svc *service.UserService, emailSvc *service.EmailService) *UserHandler {
	return &UserHandler{svc: svc, emailSvc: emailSvc}
}

func translateBindingError(err error) string {
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		for _, e := range verr {
			switch e.Field() {
			case "Username":
				if e.Tag() == "min" {
					return "用户名至少3个字符"
				}
				if e.Tag() == "max" {
					return "用户名不超过50个字符"
				}
				return "用户名格式不正确"
			case "Password":
				if e.Tag() == "min" {
					return "密码长度至少8位"
				}
				return "密码格式不正确"
			case "ConfirmPassword":
				return "两次密码输入不一致"
			case "Email":
				return "邮箱格式不正确"
			case "OldPassword":
				return "请输入原密码"
			case "NewPassword":
				if e.Tag() == "min" {
					return "新密码长度至少8位"
				}
				return "新密码格式不正确"
			case "Nickname":
				return "昵称不超过50个字符"
			case "Avatar":
				return "头像地址过长"
			case "Code":
				return "请输入 6 位数字验证码"
			case "Purpose":
				return "purpose 必须为 verify 或 change"
			case "InviteCode":
				return "邀请码格式不正确"
			}
		}
	}
	msg := err.Error()
	if strings.Contains(msg, "invalid JSON") {
		return "请求数据格式错误"
	}
	return "参数错误"
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	resp, err := h.svc.Register(&req)
	if err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	resp, err := h.svc.Login(&req)
	if err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")
	info, err := h.svc.GetUserInfo(userID)
	if err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, info)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	userID := c.GetUint("user_id")
	if err := h.svc.UpdateProfile(userID, &req); err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	userID := c.GetUint("user_id")
	if err := h.svc.ChangePassword(userID, &req); err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, nil)
}

// SendEmailCode 发送邮箱验证码（公开接口，带限流）
func (h *UserHandler) SendEmailCode(c *gin.Context) {
	var req dto.SendEmailCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Error(c, 10003, "请先登录")
		return
	}
	if err := h.emailSvc.SendCode(userID, req.Email, req.Purpose); err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, nil)
}

// VerifyEmailCode 校验邮箱验证码（鉴权接口）
func (h *UserHandler) VerifyEmailCode(c *gin.Context) {
	var req dto.VerifyEmailCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	userID := c.GetUint("user_id")
	if err := h.emailSvc.VerifyAndBind(userID, req.Email, req.Code, req.Purpose); err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, nil)
}

// SendResetPasswordCode 未登录场景发送重置密码验证码（公开接口，带限流）
//
// 安全策略：邮箱不存在 / 未验证时也返回成功（防枚举）
func (h *UserHandler) SendResetPasswordCode(c *gin.Context) {
	var req dto.ResetPasswordSendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	if err := h.emailSvc.SendResetPasswordCode(req.Email); err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, nil)
}

// ResetPasswordByCode 未登录场景校验验证码并重置密码（公开接口，带限流）
func (h *UserHandler) ResetPasswordByCode(c *gin.Context) {
	var req dto.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 10002, translateBindingError(err))
		return
	}
	if err := h.emailSvc.ResetPasswordByCode(req.Email, req.Code, req.NewPassword); err != nil {
		response.Error(c, 10001, err.Error())
		return
	}
	response.Success(c, nil)
}
