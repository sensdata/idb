package utils

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api64.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ipBytes)), nil
}

func GetPrivateIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil || ipNet.IP.IsLoopback() {
				continue
			}
			ip := ipNet.IP.To4()
			if ip != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no private IP found")
}

func GetMACAddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback == 0 && len(iface.HardwareAddr) != 0 {
			return iface.HardwareAddr.String(), nil
		}
	}
	return "", fmt.Errorf("no valid MAC address found")
}

func GenerateFingerprint(ip, mac string) string {
	data := ip + "|" + mac
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func VerifyLicenseSignature(licenseB64 string, signatureB64 string, publicKeyB64 []byte) error {
	pubBytes, err := DecodeEd25519PublicKey(publicKeyB64)
	if err != nil {
		return err
	}

	licenseBytes, err := base64.StdEncoding.DecodeString(licenseB64)
	if err != nil {
		return err
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return err
	}

	if !ed25519.Verify(pubBytes, licenseBytes, sigBytes) {
		return errors.New("license signature verification failed")
	}
	return nil
}

func DecodeEd25519PublicKey(b64pem []byte) ([]byte, error) {
	// 先 base64 decode
	pemBytes, err := base64.StdEncoding.DecodeString(string(b64pem))
	if err != nil {
		return nil, err
	}

	// 解析 PEM
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	// 从 PKIX DER 转 Ed25519 公钥
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not Ed25519 public key")
	}

	return pubKey, nil
}
