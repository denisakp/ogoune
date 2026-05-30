package crypto

// EncryptCredentialPassword encrypts a ResourceCredential.Password payload.
// On-disk format is byte-identical to Encrypt.
func EncryptCredentialPassword(plaintext string) (string, error) {
	return Encrypt(plaintext)
}

// DecryptCredentialPassword is the inverse of EncryptCredentialPassword.
func DecryptCredentialPassword(ciphertext string) (string, error) {
	return Decrypt(ciphertext)
}

// EncryptCredentialOptions encrypts a ResourceCredential.Options payload.
// On-disk format is byte-identical to Encrypt.
func EncryptCredentialOptions(plaintext string) (string, error) {
	return Encrypt(plaintext)
}

// DecryptCredentialOptions is the inverse of EncryptCredentialOptions.
func DecryptCredentialOptions(ciphertext string) (string, error) {
	return Decrypt(ciphertext)
}
