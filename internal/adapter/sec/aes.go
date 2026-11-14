// Package sec 实现 AES-GCM、租户密钥与 JWT 适配器。
package sec

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

const (
	nonceSize = 12
	tagSize   = 16
)

// AESEncrypt 生成 nonce||ciphertext||tag，密钥不足 32 字节时 SHA-256 派生。
func AESEncrypt(plain, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	out := make([]byte, nonceSize+len(plain)+tagSize)
	copy(out, nonce)
	gcm.Seal(out[nonceSize:nonceSize], nonce, plain, nil)
	return out, nil
}

// AESDecrypt 解密密文，格式错误或认证失败返回错误。
func AESDecrypt(sealed, key []byte) ([]byte, error) {
	if len(sealed) < nonceSize+tagSize {
		return nil, errors.New("密文格式不正确")
	}
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return nil, err
	}
	nonce := sealed[:nonceSize]
	return gcm.Open(nil, nonce, sealed[nonceSize:], nil)
}

func deriveKey(key []byte) []byte {
	if len(key) == 32 {
		return key
	}
	sum := sha256.Sum256(key)
	return sum[:]
}

// Encryptor 使用租户当前数据密钥加解密。
type Encryptor struct {
	keys port.KeyStore
}

// NewEncryptor 构造租户加解密器。
func NewEncryptor(keys port.KeyStore) *Encryptor { return &Encryptor{keys: keys} }

// Encrypt 使用租户当前数据密钥加密明文。
func (e *Encryptor) Encrypt(ctx context.Context, tenantID uuid.UUID, plaintext []byte) ([]byte, error) {
	key, err := e.keys.CurrentKey(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out, err := AESEncrypt(plaintext, key)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "特征加密失败")
	}
	return out, nil
}

// Decrypt 使用租户当前数据密钥解密；失败不降级明文。
func (e *Encryptor) Decrypt(ctx context.Context, tenantID uuid.UUID, cipherText []byte) ([]byte, error) {
	key, err := e.keys.CurrentKey(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	plain, err := AESDecrypt(cipherText, key)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "特征密文解密失败（密钥不匹配），不降级明文")
	}
	return plain, nil
}

// KeyStore 落库租户密钥：数据密钥由主密钥包裹。
type KeyStore struct {
	repo   port.KeyRepo
	master []byte
}

// NewKeyStore 解析 Base64 主密钥并构造密钥库；主密钥为空时延迟到使用时失败。
func NewKeyStore(repo port.KeyRepo, masterBase64 string) (*KeyStore, error) {
	if masterBase64 == "" {
		return &KeyStore{repo: repo}, nil
	}
	raw, err := base64.StdEncoding.DecodeString(masterBase64)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "SECURITY_MASTER_KEY 必须是 Base64")
	}
	return &KeyStore{repo: repo, master: raw}, nil
}

// CurrentKey 返回租户当前数据密钥；不存在时生成并原子写入。
func (k *KeyStore) CurrentKey(ctx context.Context, tenantID uuid.UUID) ([]byte, error) {
	if len(k.master) == 0 {
		return nil, domain.New(domain.CodeServiceUnavailable, "未配置 SECURITY_MASTER_KEY")
	}
	if cipher, _, err := k.repo.FindActive(ctx, tenantID); err == nil && len(cipher) > 0 {
		return k.decrypt(cipher)
	}
	maxVersion, err := k.repo.MaxVersion(ctx, tenantID)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "密钥库不可用")
	}
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "生成租户密钥失败")
	}
	wrapped, err := AESEncrypt(dek, k.master)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "包裹租户密钥失败")
	}
	_ = k.repo.Insert(ctx, tenantID, maxVersion+1, wrapped)

	cipher, _, err := k.repo.FindActive(ctx, tenantID)
	if err != nil || len(cipher) == 0 {
		return nil, domain.New(domain.CodeServiceUnavailable, "密钥库不可用，拒绝处理人脸特征")
	}
	return k.decrypt(cipher)
}

func (k *KeyStore) decrypt(cipherText []byte) ([]byte, error) {
	key, err := AESDecrypt(cipherText, k.master)
	if err != nil {
		return nil, domain.New(domain.CodeServiceUnavailable, "密钥库不可用，拒绝处理人脸特征")
	}
	return key, nil
}
