package crypto

// EncryptChannelConfig encrypts a NotificationChannel.Config payload.
//
// This is a typed indirection over the generic Encrypt primitive: the
// on-disk format is byte-identical, but per-concern entry points make it
// easier for future sqlc-backed callers (and audit) to reason about
// which payloads pass through which encryption boundary.
func EncryptChannelConfig(plaintext string) (string, error) {
	return Encrypt(plaintext)
}

// DecryptChannelConfig is the inverse of EncryptChannelConfig.
// Error semantics are identical to Decrypt — callers comparing via
// errors.Is against package sentinels see no change.
func DecryptChannelConfig(ciphertext string) (string, error) {
	return Decrypt(ciphertext)
}
