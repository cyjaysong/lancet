// Copyright 2021 dudaodong@gmail.com. All rights reserved.
// Use of this source code is governed by MIT license

// Package cryptor implements some util functions to encrypt and decrypt.
// Note:
// 1. for aes crypt function, the `key` param length should be 16, 24 or 32. if not, will panic.
package cryptor

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
	"strings"
)

// AesEcbEncrypt encrypt data with key use AES ECB algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/jT5irszHx-j
func AesEcbEncrypt(data, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	length := (len(data) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)

	copy(plain, data)

	pad := byte(len(plain) - len(data))
	for i := len(data); i < len(plain); i++ {
		plain[i] = pad
	}

	encrypted := make([]byte, len(plain))
	cipher, _ := aes.NewCipher(generateAesKey(key, size))

	for bs, be := 0, cipher.BlockSize(); bs <= len(data); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted
}

// AesEcbDecrypt decrypt data with key use AES ECB algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/jT5irszHx-j
func AesEcbDecrypt(encrypted, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}
	cipher, _ := aes.NewCipher(generateAesKey(key, size))
	decrypted := make([]byte, len(encrypted))

	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim]
}

// AesCbcEncrypt encrypt data with key use AES CBC algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/IOq_g8_lKZD
func AesCbcEncrypt(data, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	block, _ := aes.NewCipher(key)
	data = pkcs7Padding(data, block.BlockSize())

	encrypted := make([]byte, aes.BlockSize+len(data))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted[aes.BlockSize:], data)

	return encrypted
}

// AesCbcDecrypt decrypt data with key use AES CBC algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/IOq_g8_lKZD
func AesCbcDecrypt(encrypted, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	block, _ := aes.NewCipher(key)

	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encrypted, encrypted)

	decrypted := pkcs7UnPadding(encrypted)
	return decrypted
}

// AesCtrCrypt encrypt data with key use AES CTR algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/SpaZO0-5Nsp
func AesCtrCrypt(data, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	block, _ := aes.NewCipher(key)

	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)

	dst := make([]byte, len(data))
	stream.XORKeyStream(dst, data)

	return dst
}

// AesCfbEncrypt encrypt data with key use AES CFB algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/tfkF10B13kH
func AesCfbEncrypt(data, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	encrypted := make([]byte, aes.BlockSize+len(data))
	iv := encrypted[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], data)

	return encrypted
}

// AesCfbDecrypt decrypt data with key use AES CFB algorithm
// len(encrypted) should be great than 16, len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/tfkF10B13kH
func AesCfbDecrypt(encrypted, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	if len(encrypted) < aes.BlockSize {
		panic("encrypted data is too short")
	}

	block, _ := aes.NewCipher(key)
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(encrypted, encrypted)

	return encrypted
}

// AesOfbEncrypt encrypt data with key use AES OFB algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/VtHxtkUj-3F
func AesOfbEncrypt(data, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	data = pkcs7Padding(data, aes.BlockSize)
	encrypted := make([]byte, aes.BlockSize+len(data))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], data)

	return encrypted
}

// AesOfbDecrypt decrypt data with key use AES OFB algorithm
// len(key) should be 16, 24 or 32.
// Play: https://go.dev/play/p/VtHxtkUj-3F
func AesOfbDecrypt(data, key []byte) []byte {
	size := len(key)
	if size != 16 && size != 24 && size != 32 {
		panic("key length shoud be 16 or 24 or 32")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	if len(data)%aes.BlockSize != 0 {
		return nil
	}

	decrypted := make([]byte, len(data))
	mode := cipher.NewOFB(block, iv)
	mode.XORKeyStream(decrypted, data)

	decrypted = pkcs7UnPadding(decrypted)

	return decrypted
}

// AesGcmEncrypt encrypt data with key use AES GCM algorithm
// Play: https://go.dev/play/p/rUt0-DmsPCs
func AesGcmEncrypt(data, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext
}

// AesGcmDecrypt decrypt data with key use AES GCM algorithm
// Play: https://go.dev/play/p/rUt0-DmsPCs
func AesGcmDecrypt(data, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		panic("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	return plaintext
}

// DesEcbEncrypt encrypt data with key use DES ECB algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/8qivmPeZy4P
func DesEcbEncrypt(data, key []byte) []byte {
	length := (len(data) + des.BlockSize) / des.BlockSize
	plain := make([]byte, length*des.BlockSize)
	copy(plain, data)

	pad := byte(len(plain) - len(data))
	for i := len(data); i < len(plain); i++ {
		plain[i] = pad
	}

	encrypted := make([]byte, len(plain))
	cipher, _ := des.NewCipher(generateDesKey(key))

	for bs, be := 0, cipher.BlockSize(); bs <= len(data); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted
}

// DesEcbDecrypt decrypt data with key use DES ECB algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/8qivmPeZy4P
func DesEcbDecrypt(encrypted, key []byte) []byte {
	cipher, _ := des.NewCipher(generateDesKey(key))
	decrypted := make([]byte, len(encrypted))

	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim]
}

// DesCbcEncrypt encrypt data with key use DES CBC algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/4cC4QvWfe3_1
func DesCbcEncrypt(data, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, _ := des.NewCipher(key)
	data = pkcs7Padding(data, block.BlockSize())

	encrypted := make([]byte, des.BlockSize+len(data))
	iv := encrypted[:des.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted[des.BlockSize:], data)

	return encrypted
}

// DesCbcDecrypt decrypt data with key use DES CBC algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/4cC4QvWfe3_1
func DesCbcDecrypt(encrypted, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, _ := des.NewCipher(key)

	iv := encrypted[:des.BlockSize]
	encrypted = encrypted[des.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encrypted, encrypted)

	decrypted := pkcs7UnPadding(encrypted)
	return decrypted
}

// DesCtrCrypt encrypt data with key use DES CTR algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/9-T6OjKpcdw
func DesCtrCrypt(data, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, _ := des.NewCipher(key)

	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)

	dst := make([]byte, len(data))
	stream.XORKeyStream(dst, data)

	return dst
}

// DesCfbEncrypt encrypt data with key use DES CFB algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/y-eNxcFBlxL
func DesCfbEncrypt(data, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic(err)
	}

	encrypted := make([]byte, des.BlockSize+len(data))
	iv := encrypted[:des.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[des.BlockSize:], data)

	return encrypted
}

// DesCfbDecrypt decrypt data with key use DES CFB algorithm
// len(encrypted) should be great than 16, len(key) should be 8.
// Play: https://go.dev/play/p/y-eNxcFBlxL
func DesCfbDecrypt(encrypted, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, _ := des.NewCipher(key)
	if len(encrypted) < des.BlockSize {
		panic("encrypted data is too short")
	}
	iv := encrypted[:des.BlockSize]
	encrypted = encrypted[des.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)

	return encrypted
}

// DesOfbEncrypt encrypt data with key use DES OFB algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/74KmNadjN1J
func DesOfbEncrypt(data, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic(err)
	}
	data = pkcs7Padding(data, des.BlockSize)
	encrypted := make([]byte, des.BlockSize+len(data))
	iv := encrypted[:des.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(encrypted[des.BlockSize:], data)

	return encrypted
}

// DesOfbDecrypt decrypt data with key use DES OFB algorithm
// len(key) should be 8.
// Play: https://go.dev/play/p/74KmNadjN1J
func DesOfbDecrypt(data, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic(err)
	}

	iv := data[:des.BlockSize]
	data = data[des.BlockSize:]
	if len(data)%des.BlockSize != 0 {
		return nil
	}

	decrypted := make([]byte, len(data))
	mode := cipher.NewOFB(block, iv)
	mode.XORKeyStream(decrypted, data)

	decrypted = pkcs7UnPadding(decrypted)

	return decrypted
}

// GenerateRsaKey create rsa private and public pemo file.
// Play: https://go.dev/play/p/zutRHrDqs0X
func GenerateRsaKey(keySize int, priKeyFile, pubKeyFile string) error {
	// private key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return err
	}

	derText := x509.MarshalPKCS1PrivateKey(privateKey)

	block := pem.Block{
		Type:  "rsa private key",
		Bytes: derText,
	}

	file, err := os.Create(priKeyFile)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(file, &block)
	if err != nil {
		return err
	}

	file.Close()

	// public key
	publicKey := privateKey.PublicKey

	derpText, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}

	block = pem.Block{
		Type:  "rsa public key",
		Bytes: derpText,
	}

	file, err = os.Create(pubKeyFile)
	if err != nil {
		return err
	}

	err = pem.Encode(file, &block)
	if err != nil {
		return err
	}

	file.Close()

	return nil
}

// RsaEncrypt encrypt data with ras algorithm.
// Play: https://go.dev/play/p/7_zo6mrx-eX
func RsaEncrypt(data []byte, pubKeyFileName string) []byte {
	file, err := os.Open(pubKeyFileName)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf := make([]byte, fileInfo.Size())

	_, err = file.Read(buf)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(buf)

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	pubKey := pubInterface.(*rsa.PublicKey)

	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		panic(err)
	}

	return cipherText
}

// RsaDecrypt decrypt data with ras algorithm.
// Play: https://go.dev/play/p/7_zo6mrx-eX
func RsaDecrypt(data []byte, privateKeyFileName string) []byte {
	file, err := os.Open(privateKeyFileName)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, fileInfo.Size())
	defer file.Close()

	_, err = file.Read(buf)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(buf)

	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, priKey, data)
	if err != nil {
		panic(err)
	}

	return plainText
}

// GenerateRsaKeyPair create rsa private and public key.
// Play: https://go.dev/play/p/sSVmkfENKMz
func GenerateRsaKeyPair(keySize int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, keySize)
	return privateKey, &privateKey.PublicKey
}

// RsaEncryptOAEP encrypts the given data with RSA-OAEP.
// Play: https://go.dev/play/p/sSVmkfENKMz
func RsaEncryptOAEP(data []byte, label []byte, key rsa.PublicKey) ([]byte, error) {
	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key, data, label)
	if err != nil {
		return nil, err
	}

	return encryptedBytes, nil
}

// RsaDecryptOAEP decrypts the data with RSA-OAEP.
// Play: https://go.dev/play/p/sSVmkfENKMz
func RsaDecryptOAEP(ciphertext []byte, label []byte, key rsa.PrivateKey) ([]byte, error) {
	decryptedBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, &key, ciphertext, label)
	if err != nil {
		return nil, err
	}

	return decryptedBytes, nil
}

// RsaSign signs the data with RSA.
// Play: https://go.dev/play/p/qhsbf8BJ6Mf
func RsaSign(hash crypto.Hash, data []byte, privateKeyFileName string) ([]byte, error) {
	privateKey, err := loadRasPrivateKey(privateKeyFileName)
	if err != nil {
		return nil, err
	}

	hashed, err := hashData(hash, data)
	if err != nil {
		return nil, err
	}

	return rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
}

// RsaVerifySign verifies the signature of the data with RSA.
// Play: https://go.dev/play/p/qhsbf8BJ6Mf
func RsaVerifySign(hash crypto.Hash, data, signature []byte, pubKeyFileName string) error {
	publicKey, err := loadRsaPublicKey(pubKeyFileName)
	if err != nil {
		return err
	}

	hashed, err := hashData(hash, data)
	if err != nil {
		return err
	}

	return rsa.VerifyPKCS1v15(publicKey, hash, hashed, signature)
}

// loadRsaPrivateKey loads and parses a PEM encoded private key file.
func loadRsaPublicKey(filename string) (*rsa.PublicKey, error) {
	pubKeyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pubKeyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing the public key")
	}

	var pubKey *rsa.PublicKey
	blockType := strings.ToUpper(block.Type)

	if blockType == "RSA PUBLIC KEY" {
		pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			key, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}

			var ok bool
			pubKey, ok = key.(*rsa.PublicKey)
			if !ok {
				return nil, errors.New("failed to parse RSA private key")
			}
		}
	} else if blockType == "PUBLIC KEY" {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		var ok bool
		pubKey, ok = key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("failed to parse RSA private key")
		}

	} else {
		return nil, errors.New("unsupported key type")
	}

	return pubKey, nil
}

// loadRsaPrivateKey loads and parses a PEM encoded private key file.
func loadRasPrivateKey(filename string) (*rsa.PrivateKey, error) {
	priKeyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(priKeyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing the private key")
	}

	var privateKey *rsa.PrivateKey
	blockType := strings.ToUpper(block.Type)

	// PKCS#1 format
	if blockType == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	} else if blockType == "PRIVATE KEY" { // PKCS#8 format
		priKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privateKey, ok = priKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("failed to parse RSA private key")
		}
	} else {
		return nil, errors.New("unsupported key type")
	}

	return privateKey, nil
}

// hashData returns the hash value of the data, using the specified hash function
func hashData(hash crypto.Hash, data []byte) ([]byte, error) {
	if !hash.Available() {
		return nil, errors.New("unsupported hash algorithm")
	}

	var hashed []byte

	switch hash {
	case crypto.SHA224:
		h := sha256.Sum224(data)
		hashed = h[:]
	case crypto.SHA256:
		h := sha256.Sum256(data)
		hashed = h[:]
	case crypto.SHA384:
		h := sha512.Sum384(data)
		hashed = h[:]
	case crypto.SHA512:
		h := sha512.Sum512(data)
		hashed = h[:]
	default:
		return nil, errors.New("unsupported hash algorithm")
	}

	return hashed, nil
}
