package encryption

import (
	"bytes"
	"encoding/base64"
	"log"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/testing/fakekms"
	"github.com/tink-crypto/tink-go/v2/tink"
)

var CryptoPrimitive tink.AEAD
var associatedData = []byte("example encryption")

// الفانكشن الجديدة التي طلبتها لإعداد التشفير وتعبئة المتغير العام
func InitCrypto(keyURI string) {
	client, err := fakekms.NewClient(keyURI)
	if err != nil {
		log.Fatal(err)
	}
	kekAEAD, err := client.GetAEAD(keyURI)
	if err != nil {
		log.Fatal(err)
	}

	newHandle, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	writer := keyset.NewBinaryWriter(buf)
	err = newHandle.WriteWithAssociatedData(writer, kekAEAD, []byte("keyset encryption example"))
	if err != nil {
		log.Fatal(err)
	}

	reader := keyset.NewBinaryReader(bytes.NewReader(buf.Bytes()))
	handle, err := keyset.ReadWithAssociatedData(reader, kekAEAD, []byte("keyset encryption example"))
	if err != nil {
		log.Fatal(err)
	}

	// حقن القيمة مباشرة في المتغير العام داخل الحزمة
	CryptoPrimitive, err = aead.New(handle)
	if err != nil {
		log.Fatalf("فشل تعبئة المتغير العام: %v", err)
	}
}

// دالة التشفير
func EncryptText(plainText string) (string, error) {
	cipherBytes, err := CryptoPrimitive.Encrypt([]byte(plainText), associatedData)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

// دالة فك التشفير
func DecryptText(cipherTextBase64 string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}
	plainBytes, err := CryptoPrimitive.Decrypt(cipherBytes, associatedData)
	if err != nil {
		return "", err
	}
	return string(plainBytes), nil
}
