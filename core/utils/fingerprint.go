package utils

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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

func VerifyLicenseSignature(license string, signature string, publicKeyBase64 string) error {
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return err
	}

	// 编码 license 对象为 JSON
	licenseJSON, err := json.Marshal(license)
	if err != nil {
		return err
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	if !ed25519.Verify(pubBytes, licenseJSON, sigBytes) {
		return errors.New("license signature verification failed")
	}

	return nil
}
