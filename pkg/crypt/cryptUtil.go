package cryptUtil

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptCipher 实现 Cipher 接口，使用 bcrypt 进行密码加密和验证
type BcryptCipher struct {
	// Cost 是 bcrypt 的计算强度，建议值在 10-14 之间
	Cost int
}

// NewBcryptCipher 创建一个新的 BcryptCipher 实例
func NewBcryptCipher(cost int) *BcryptCipher {
	// 确保 cost 在合理范围内（bcrypt 默认最小值为 4，最大值为 31）
	if cost < 4 || cost > 31 {
		cost = bcrypt.DefaultCost // 使用默认值 10
	}
	return &BcryptCipher{Cost: cost}
}

// Encrypt 把明文加密成密文（bcrypt 哈希）
func (c *BcryptCipher) Encrypt(plaintext string) (encrypted string, err error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), c.Cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Verify 检查密文是否与明文匹配
func (c *BcryptCipher) Verify(ciphertext string, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(ciphertext), []byte(plaintext))
}
