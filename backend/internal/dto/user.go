package dto

type RegisterReq struct {
	Username        string `json:"username" binding:"required,min=3,max=50"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	Email           string `json:"email" binding:"required,email"`
	LevelID         uint   `json:"level_id" binding:"required,gt=0"`
	SubjectID       uint   `json:"subject_id" binding:"required,gt=0"`
	Difficulty      string `json:"difficulty"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	AccessToken string   `json:"access_token"`
	ExpiresIn   int      `json:"expires_in"`
	UserInfo    UserInfo `json:"user_info"`
}

type UserInfo struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Role        string `json:"role"`
	LevelID     uint   `json:"level_id"`
	SubjectID   uint   `json:"subject_id"`
	Difficulty  string `json:"difficulty"`
	LevelName   string `json:"level_name"`
	SubjectName string `json:"subject_name"`
}

type UpdateProfileReq struct {
	Nickname   string  `json:"nickname" binding:"max=50"`
	Avatar     string  `json:"avatar" binding:"max=255"`
	LevelID    *uint   `json:"level_id"`
	SubjectID  *uint   `json:"subject_id"`
	Difficulty *string `json:"difficulty"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
