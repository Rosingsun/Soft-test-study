import { get, post, put } from './request'
import type { LoginReq, RegisterReq, LoginResp, UserInfo, UpdateProfileReq, ChangePasswordReq } from '@/types/user'

export function login(data: LoginReq) {
  return post<LoginResp>('/auth/login', data)
}

export function register(data: RegisterReq) {
  return post<null>('/auth/register', data)
}

export function getUserInfo() {
  return get<UserInfo>('/auth/user-info')
}

export function updateProfile(data: UpdateProfileReq) {
  return put<null>('/auth/profile', data)
}

export function changePassword(data: ChangePasswordReq) {
  return put<null>('/auth/password', data)
}
