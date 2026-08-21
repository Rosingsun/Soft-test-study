// Package security 提供密码哈希与校验工具，封装 bcrypt 细节。
// 后续 OPT-11 调整 cost 时，仅需改本文件即可全局生效。
package security

import "golang.org/x/crypto/bcrypt"

// HashPassword 用 bcrypt 对明文密码进行哈希。
// 失败原因仅可能是 bcrypt 内部错误（极少见，常见原因是密码超长被截断）。
func HashPassword(p string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// VerifyPassword 校验明文密码与 bcrypt 哈希是否匹配。
// 匹配返回 nil；不匹配或 hash 非法返回 bcrypt.ErrMismatchedHashAndPassword 等。
func VerifyPassword(hash, p string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
}
