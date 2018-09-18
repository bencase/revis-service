package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const keyFilename = "key.txt"
const defaultHexStr = "f35116b13bd7345ff8f559c854a9b2accc108451bac36040400f06298b08e8b8"

func getKey() (*[32]byte, error) {
	// Check if the key file exists
	_, err := os.Stat(keyFilename)
	fileExists := !os.IsNotExist(err)
	var hexStr string
	if fileExists {
		hexStr, err = getHexStrFromFile()
		if err != nil { return nil, err }
	} else {
		hexStr = getDefaultHexStr()
	}

	decodedBytes, err := hex.DecodeString(hexStr)
	if err != nil { return nil, err }
	if len(decodedBytes) < 32 { return nil, errors.New("Key " + hexStr + " is not long enough") }
	key := [32]byte{}
	_, err = io.ReadFull(bytes.NewReader(decodedBytes), key[:])
	if err != nil { return nil, err }
	return &key, nil
}
func getHexStrFromFile() (string, error) {
	fileBytes, err := ioutil.ReadFile(keyFilename)
	if err != nil { return "", err }

	hexStr := string(fileBytes)
	return strings.TrimSpace(hexStr), nil
}
func getDefaultHexStr() string {
	return defaultHexStr
}

func EncryptToBase64(text string) (string, error) {
	ciphertext, err := Encrypt(text)
	if err != nil { return "", err }
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
func Encrypt(text string) ([]byte, error) {
	key, err := getKey()
	if err != nil { return nil, err }

	block, err := aes.NewCipher(key[:])
	if err != nil { return nil, err }
	
	gcm, err := cipher.NewGCM(block)
	if err != nil { return nil, err }

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil { return nil, err }

	return gcm.Seal(nonce, nonce, []byte(text), nil), nil
}

func DecryptFromBase64(b64Text string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(b64Text)
	if err != nil { return "", err }
	return Decrypt(decodedBytes)
}
func Decrypt(ciphertext []byte) (string, error) {
	key, err := getKey()
	if err != nil { return "", err }

	block, err := aes.NewCipher(key[:])
	if err != nil { return "", err }
	
	gcm, err := cipher.NewGCM(block)
	if err != nil { return "", err }

	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("Malformed ciphertext")
	}

	decryptedBytes, err := gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
	if err != nil { return "", err }

	return string(decryptedBytes), nil
}