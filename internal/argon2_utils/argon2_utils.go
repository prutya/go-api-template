package argon2_utils

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidHash = errors.New("the encoded hash is not in the correct format")
var ErrIncompatibleVersion = errors.New("incompatible version of argon2")

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
}

func CalculateAndEncode(data []byte, salt []byte, params *Params) string {
	return Encode(salt, Calculate(data, salt, params), params)
}

func Calculate(data []byte, salt []byte, params *Params) []byte {
	return argon2.IDKey(
		data,
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)
}

func Encode(
	salt []byte,
	hash []byte,
	params *Params,
) string {
	// Base64 encode the salt and hashed OTP.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	)
}

func Compare(plaintext string, encodedHash string) (bool, error) {
	// Extract the parameters, salt and derived key from the encoded hash
	params, salt, hash, err := Decode(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the plaintext using the same parameters.
	otherHash := Calculate([]byte(plaintext), salt, params)

	// Check that the contents of the hashed values are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func Decode(encodedHash string) (*Params, []byte, []byte, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(vals[2], "v=%d", &version); err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p := &Params{}
	if _, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism); err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	// #nosec G115 -- salt length is guaranteed to be < 2^32
	// p.saltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	// #nosec G115 -- key length is guaranteed to be < 2^32
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
