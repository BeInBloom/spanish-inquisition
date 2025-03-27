package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	password = "x35k9f"
	msg      = `0ba7cd8c624345451df4710b81d1a349ce401e61bc7eb704ca` +
		`a84a8cde9f9959699f75d0d1075d676f1fe2eb475cf81f62ef` +
		`f701fee6a433cfd289d231440cf549e40b6c13d8843197a95f` +
		`8639911b7ed39a3aec4dfa9d286095c705e1a825b10a9104c6` +
		`be55d1079e6c6167118ac91318fe`
)

func main() {
	// 1) Получаем ключ из password, используя sha256.Sum256
	key := sha256.Sum256([]byte(password))

	// 2) Создаем aesblock и aesgcm
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		panic(err)
	}

	// 3) Декодируем сообщение msg в двоичный формат
	b, err := hex.DecodeString(msg)
	if err != nil {
		panic(err)
	}

	// 4) Получаем вектор инициализации (nonce) из начала зашифрованных данных
	nonceSize := aesgcm.NonceSize()
	if len(b) < nonceSize {
		panic("зашифрованные данные слишком короткие")
	}

	nonce, ciphertext := b[:nonceSize], b[nonceSize:]

	// 5) Расшифровываем и выводим данные
	text, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(text))
}
