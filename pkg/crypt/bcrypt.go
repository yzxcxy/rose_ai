package cryptUtil

// Cipher 定义了密码加密和验证的接口
type Cipher interface {
	// Encrypt 将明文加密为密文
	// 参数:
	//   plaintext: 需要加密的明文字符串
	// 返回:
	//   encrypted: 加密后的密文字符串
	//   err: 加密过程中可能发生的错误
	Encrypt(plaintext string) (encrypted string, err error)

	// Verify 检查密文是否与明文匹配
	// 参数:
	//   ciphertext: 存储的密文字符串
	//   plaintext: 需要验证的明文字符串
	// 返回:
	//   err: 如果匹配则返回 nil，否则返回错误
	Verify(ciphertext string, plaintext string) (err error)
}
