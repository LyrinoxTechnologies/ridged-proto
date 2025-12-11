package rdgproto

import (
"crypto"
"crypto/hmac"
"crypto/rand"
"crypto/rsa"
"crypto/sha256"
)

// HMACSigner implements HMAC-SHA256 signing
type HMACSigner struct {
secret []byte
}

// NewHMACSigner creates a new HMAC signer
func NewHMACSigner(secret []byte) *HMACSigner {
return &HMACSigner{secret: secret}
}

// Sign creates an HMAC-SHA256 signature
func (h *HMACSigner) Sign(data []byte) ([]byte, error) {
mac := hmac.New(sha256.New, h.secret)
mac.Write(data)
return mac.Sum(nil), nil
}

// HMACVerifier implements HMAC-SHA256 verification
type HMACVerifier struct {
secret []byte
}

// NewHMACVerifier creates a new HMAC verifier
func NewHMACVerifier(secret []byte) *HMACVerifier {
return &HMACVerifier{secret: secret}
}

// Verify checks an HMAC-SHA256 signature
func (h *HMACVerifier) Verify(data []byte, signature []byte) error {
mac := hmac.New(sha256.New, h.secret)
mac.Write(data)
expected := mac.Sum(nil)
if !hmac.Equal(expected, signature) {
return ErrInvalidSignature
}
return nil
}

// SecureMessageOptions creates MessageOptions with HMAC signing enabled
func SecureMessageOptions(secret []byte) *MessageOptions {
return &MessageOptions{
Signer:   NewHMACSigner(secret),
Verifier: NewHMACVerifier(secret),
}
}

// RSASigner implements RSA-SHA256 signing
type RSASigner struct {
privateKey *rsa.PrivateKey
}

// NewRSASigner creates a new RSA signer
func NewRSASigner(privateKey *rsa.PrivateKey) *RSASigner {
return &RSASigner{privateKey: privateKey}
}

// Sign creates an RSA-SHA256 signature
func (r *RSASigner) Sign(data []byte) ([]byte, error) {
hash := sha256.Sum256(data)
return rsa.SignPKCS1v15(rand.Reader, r.privateKey, crypto.SHA256, hash[:])
}

// RSAVerifier implements RSA-SHA256 verification
type RSAVerifier struct {
publicKey *rsa.PublicKey
}

// NewRSAVerifier creates a new RSA verifier
func NewRSAVerifier(publicKey *rsa.PublicKey) *RSAVerifier {
return &RSAVerifier{publicKey: publicKey}
}

// Verify checks an RSA-SHA256 signature
func (r *RSAVerifier) Verify(data []byte, signature []byte) error {
hash := sha256.Sum256(data)
err := rsa.VerifyPKCS1v15(r.publicKey, crypto.SHA256, hash[:], signature)
if err != nil {
return ErrInvalidSignature
}
return nil
}

// RSAMessageOptions creates MessageOptions with RSA signing enabled
func RSAMessageOptions(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *MessageOptions {
opts := &MessageOptions{}
if privateKey != nil {
opts.Signer = NewRSASigner(privateKey)
}
if publicKey != nil {
opts.Verifier = NewRSAVerifier(publicKey)
}
return opts
}

// GenerateRSAKeyPair generates a new RSA key pair
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
privateKey, err := rsa.GenerateKey(rand.Reader, bits)
if err != nil {
return nil, nil, err
}
return privateKey, &privateKey.PublicKey, nil
}
