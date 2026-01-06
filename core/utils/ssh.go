package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// NewHostKeyCallback
// - host identity is explicitly provided by caller (addr + port)
// - platform-level known_hosts
// - accept-new, reject mismatch
func NewHostKeyCallback(
	knownHostsPath string,
	hostPort string, // e.g. "10.0.0.1:22" or "[2001:db8::1]:22"
) (ssh.HostKeyCallback, error) {

	sshDir := filepath.Dir(knownHostsPath)

	// 1. Ensure directory exists
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, fmt.Errorf("create ssh dir failed: %w", err)
	}

	// 2. Ensure known_hosts file exists and is safe
	if err := ensureSecureFile(knownHostsPath); err != nil {
		return nil, err
	}

	// 3. Init base known_hosts callback
	baseCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("init known_hosts failed: %w", err)
	}

	// mismatch log path
	mismatchLog := filepath.Join(sshDir, "mismatch.log")
	if err := ensureSecureFile(mismatchLog); err != nil {
		return nil, err
	}

	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		// Use explicit host identity only
		err := baseCallback(hostPort, nil, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if !errors.As(err, &keyErr) {
			return err
		}

		// HostKey mismatch → reject but log
		if len(keyErr.Want) > 0 {
			logHostKeyMismatch(
				mismatchLog,
				hostPort,
				keyErr.Want[0].Key,
				key,
			)

			return fmt.Errorf(
				"ssh host key mismatch for %s (old=%s new=%s)",
				hostPort,
				ssh.FingerprintSHA256(keyErr.Want[0].Key),
				ssh.FingerprintSHA256(key),
			)
		}

		// Host not found → TOFU (accept-new)
		f, ferr := os.OpenFile(
			knownHostsPath,
			os.O_APPEND|os.O_WRONLY,
			0600,
		)
		if ferr != nil {
			return ferr
		}
		defer f.Close()

		line := knownhosts.Line(
			[]string{hostPort},
			key,
		)

		if _, ferr = f.WriteString(line + "\n"); ferr != nil {
			return ferr
		}

		return nil
	}, nil
}

func logHostKeyMismatch(
	logPath string,
	hostPort string,
	oldKey ssh.PublicKey,
	newKey ssh.PublicKey,
) {
	f, err := os.OpenFile(
		logPath,
		os.O_APPEND|os.O_WRONLY,
		0600,
	)
	if err != nil {
		return
	}
	defer f.Close()

	ts := time.Now().Format("2006-01-02 15:04:05")

	line := fmt.Sprintf(
		"[%s] host=%s old=%s new=%s\n",
		ts,
		hostPort,
		ssh.FingerprintSHA256(oldKey),
		ssh.FingerprintSHA256(newKey),
	)

	_, _ = f.WriteString(line)
}

func ensureSecureFile(path string) error {
	fi, err := os.Lstat(path)
	if err == nil {
		// Disallow symlink
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("insecure file (symlink): %s", path)
		}
		// Enforce permission
		return os.Chmod(path, 0600)
	}

	if !os.IsNotExist(err) {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	return f.Close()
}
