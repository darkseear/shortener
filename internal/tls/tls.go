package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/darkseear/shortener/internal/logger"
)

// CrtFile и KeyFile - имена файлов для сертификата и ключа.
// Их можно изменить, если нужно использовать другие имена.
var (
	CrtFile = "server.crt"
	KeyFile = "server.key"
)

// GenerateCerts генерация сертификатов.
func GenerateCerts() error {
	// Создаём шаблон сертификата
	// Здесь можно указать свои значения для полей Subject, IPAddresses и т.д.
	// cert - это структура x509.Certificate, которая содержит информацию о сертификате.
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// Создаём приватный RSA-ключ
	// privateKey - это структура rsa.PrivateKey, которая содержит закрытый ключ.
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	// Создаём сертификат x.509
	// certBytes - это байтовый срез, содержащий сериализованный сертификат.
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Кодируем сертификат и ключ в PEM
	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.WriteFile(CrtFile, certPEM.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(KeyFile, privateKeyPEM.Bytes(), 0600)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log.Info("Сгенерирован сертификат")
	return err
}
