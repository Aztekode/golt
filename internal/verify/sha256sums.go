package verify

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type SHA256Sums map[string][32]byte

func ParseSHA256SUMS(r io.Reader) (SHA256Sums, error) {
	sums := make(SHA256Sums)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid SHA256SUMS line: %q", line)
		}

		hashHex := fields[0]
		name := fields[1]
		name = strings.TrimPrefix(name, "*")

		b, err := hex.DecodeString(hashHex)
		if err != nil || len(b) != 32 {
			return nil, fmt.Errorf("invalid sha256 in line: %q", line)
		}

		var sum [32]byte
		copy(sum[:], b)
		sums[name] = sum
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return sums, nil
}

func SHA256File(path string) ([32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return [32]byte{}, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, err
	}

	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return sum, nil
}

type VerifyAssetOptions struct {
	PublicKeyBase64 string
	SHA256SUMSPath  string
	SignaturePath   string
}

func VerifyAsset(filePath string, opt VerifyAssetOptions) (string, error) {
	if opt.PublicKeyBase64 == "" {
		return "", errors.New("missing minisign public key")
	}

	pub, err := ParseMinisignPublicKey(opt.PublicKeyBase64)
	if err != nil {
		return "", err
	}

	shaPath := opt.SHA256SUMSPath
	if shaPath == "" {
		shaPath = filepath.Join(filepath.Dir(filePath), "SHA256SUMS")
	}

	sigPath := opt.SignaturePath
	if sigPath == "" {
		sigPath = shaPath + ".minisig"
	}

	shaBytes, err := os.ReadFile(shaPath)
	if err != nil {
		return "", err
	}

	sigBytes, err := os.ReadFile(sigPath)
	if err != nil {
		return "", err
	}

	trustedComment, err := VerifyMinisigBytes(pub, shaBytes, sigBytes)
	if err != nil {
		return "", err
	}

	sums, err := ParseSHA256SUMS(strings.NewReader(string(shaBytes)))
	if err != nil {
		return "", err
	}

	baseName := filepath.Base(filePath)
	expected, ok := sums[baseName]
	if !ok {
		return "", fmt.Errorf("file %q not found in SHA256SUMS", baseName)
	}

	actual, err := SHA256File(filePath)
	if err != nil {
		return "", err
	}

	if actual != expected {
		return "", errors.New("sha256 mismatch")
	}

	return trustedComment, nil
}
