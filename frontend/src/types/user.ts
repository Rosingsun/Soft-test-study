export interface LoginReq {
  username: string
  password: string
}

export interface RegisterReq {
  username: string
  password: string
  confirm_password: string
  email: string
  level_id?: number
  subject_id?: number
  difficulty?: string
}

export interface LoginResp {
  access_token: string
  expires_in: number
  user_info: UserInfo
}

export interface UserInfo {
  id: number
  username: string
  email: string
  nickname: string
  avatar: string
  role: string
  level_id: number
  subject_id: number
  difficulty: string
  level_name: string
  subject_name: string
}

export interface UpdateProfileReq {
  nickname?: string
  avatar?: string
  level_id?: number | null
  subject_id?: number | null
  difficulty?: string | null
}

export interface ChangePasswordReq {
  old_password: string
  new_password: string
}
