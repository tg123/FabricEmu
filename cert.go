package FabricEmu

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"math/big"
	"time"
)

func generateCert() (*tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 10 * 24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		PrivateKey:  privateKey,
		Certificate: [][]byte{cert},
	}, nil
}

func cryptExportKey(key *rsa.PrivateKey) (_ string, err error) {
	bl := key.Size() * 8

	buf := &bytes.Buffer{}

	appendBigInt := func(n *big.Int, size int) {
		if err != nil {
			return
		}

		b := n.Bytes()
		l := len(b)
		for i := l - 1; i >= 0; i-- {
			err = buf.WriteByte(b[i])
			if err != nil {
				return
			}
		}

		for i := 0; i < size-l; i++ {
			err = buf.WriteByte(b[i])
			if err != nil {
				return
			}
		}
	}

	// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-wcce/5cf2e6b9-3195-4f85-bc18-05b50e6d4e11
	err = binary.Write(buf, binary.LittleEndian, struct {
		Type     uint8
		Version  uint8
		Reserved uint16
		KeyAlg   uint32
		Magic    uint32
		Bitlen   uint32
		PubExp   uint32
		// Modulus  []byte
		// P        []byte
		// Q        []byte
		// Dp       []byte
		// Dq       []byte
		// Iq       []byte
		// D        []byte
	}{
		Type:     0x7,
		Version:  0x2,
		Reserved: 0,
		KeyAlg:   0x0000A400,
		Magic:    0x32415352,
		Bitlen:   uint32(bl),
		PubExp:   65537,
	})

	key.Precompute()

	P := key.Primes[0]
	Q := key.Primes[1]

	Modulus := new(big.Int).Mul(P, Q)

	appendBigInt(Modulus, bl/8)
	appendBigInt(P, bl/16)
	appendBigInt(Q, bl/16)
	appendBigInt(key.Precomputed.Dp, bl/16)
	appendBigInt(key.Precomputed.Dq, bl/16)
	appendBigInt(key.Precomputed.Qinv, bl/16)
	appendBigInt(key.D, bl/8)

	if err != nil {
		return
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
