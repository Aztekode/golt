package verify

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/blake2b"
)

type MinisignPublicKey struct {
	KeyID     [8]byte
	PublicKey ed25519.PublicKey
}

func ParseMinisignPublicKey(pub string) (MinisignPublicKey, error) {
	pub = strings.TrimSpace(pub)
	raw, err := base64.StdEncoding.DecodeString(pub)
	if err != nil {
		return MinisignPublicKey{}, err
	}
	if len(raw) != 2+8+ed25519.PublicKeySize {
		return MinisignPublicKey{}, fmt.Errorf("invalid minisign public key length: %d", len(raw))
	}
	if string(raw[:2]) != "Ed" {
		return MinisignPublicKey{}, fmt.Errorf("unsupported minisign key algorithm: %q", string(raw[:2]))
	}

	var keyID [8]byte
	copy(keyID[:], raw[2:10])
	pk := make([]byte, ed25519.PublicKeySize)
	copy(pk, raw[10:])

	return MinisignPublicKey{KeyID: keyID, PublicKey: ed25519.PublicKey(pk)}, nil
}

type MinisignSignature struct {
	Algorithm      string
	KeyID          [8]byte
	Signature      []byte
	TrustedComment string
	GlobalSig      []byte
}

func ParseMinisig(r io.Reader) (MinisignSignature, error) {
	sc := bufio.NewScanner(r)
	var lines []string
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil {
		return MinisignSignature{}, err
	}
	if len(lines) < 4 {
		return MinisignSignature{}, errors.New("invalid minisig: expected 4 lines")
	}

	sigLine := lines[1]
	sigRaw, err := base64.StdEncoding.DecodeString(sigLine)
	if err != nil {
		return MinisignSignature{}, err
	}
	if len(sigRaw) != 2+8+ed25519.SignatureSize {
		return MinisignSignature{}, fmt.Errorf("invalid signature length: %d", len(sigRaw))
	}

	algo := string(sigRaw[:2])
	if algo != "ED" && algo != "Ed" {
		return MinisignSignature{}, fmt.Errorf("unsupported signature algorithm: %q", algo)
	}

	var keyID [8]byte
	copy(keyID[:], sigRaw[2:10])
	sig := make([]byte, ed25519.SignatureSize)
	copy(sig, sigRaw[10:])

	trustedLine := lines[2]
	idx := strings.Index(trustedLine, ":")
	if idx < 0 {
		return MinisignSignature{}, errors.New("invalid trusted comment line")
	}
	trustedComment := strings.TrimSpace(trustedLine[idx+1:])

	globalLine := lines[3]
	globalRaw, err := base64.StdEncoding.DecodeString(globalLine)
	if err != nil {
		return MinisignSignature{}, err
	}
	if len(globalRaw) != ed25519.SignatureSize {
		return MinisignSignature{}, fmt.Errorf("invalid global signature length: %d", len(globalRaw))
	}

	globalSig := make([]byte, ed25519.SignatureSize)
	copy(globalSig, globalRaw)

	return MinisignSignature{
		Algorithm:      algo,
		KeyID:          keyID,
		Signature:      sig,
		TrustedComment: trustedComment,
		GlobalSig:      globalSig,
	}, nil
}

func VerifyMinisig(pub MinisignPublicKey, msg []byte, sig MinisignSignature) (string, error) {
	if pub.PublicKey == nil {
		return "", errors.New("missing public key")
	}
	if pub.KeyID != sig.KeyID {
		return "", errors.New("signature key id mismatch")
	}

	var signedMsg []byte
	switch sig.Algorithm {
	case "ED":
		sum := blake2b.Sum512(msg)
		signedMsg = sum[:]
	case "Ed":
		signedMsg = msg
	default:
		return "", fmt.Errorf("unsupported signature algorithm: %q", sig.Algorithm)
	}

	if !ed25519.Verify(pub.PublicKey, signedMsg, sig.Signature) {
		return "", errors.New("invalid signature")
	}

	globalMsg := append(append([]byte{}, sig.Signature...), []byte(sig.TrustedComment)...)
	if !ed25519.Verify(pub.PublicKey, globalMsg, sig.GlobalSig) {
		return "", errors.New("invalid trusted comment signature")
	}

	return sig.TrustedComment, nil
}

func VerifyMinisigBytes(pub MinisignPublicKey, msg []byte, sigFile []byte) (string, error) {
	sig, err := ParseMinisig(bytes.NewReader(sigFile))
	if err != nil {
		return "", err
	}
	return VerifyMinisig(pub, msg, sig)
}
