package idenc

import (
	"fmt"
	"math"
	"sync"

	"github.com/sqids/sqids-go"
)

// IDEncoder encodes and decodes int64 database IDs to and from opaque
// Sqids strings, hiding internal primary keys from API consumers.
type IDEncoder interface {
	// Encode returns the Sqids-encoded string for the given int64 ID.
	Encode(id int64) string
	// Decode parses an encoded Sqids string and returns the original int64 ID,
	// returning an error when the string is invalid.
	Decode(encoded string) (int64, error)
}

type sqidsEncoder struct {
	sqids *sqids.Sqids
}

// globalEncoder holds the singleton IDEncoder instance
// used by the SqID type for URI/query binding.
var (
	globalEncoder IDEncoder
	encoderOnce   sync.Once
)

// SetGlobalEncoder sets the global IDEncoder used by
// SqID.UnmarshalText. Must be called once at startup
// before any request handling.
func SetGlobalEncoder(enc IDEncoder) {
	encoderOnce.Do(func() {
		globalEncoder = enc
	})
}

// SqID is a type that wraps int64 and supports automatic
// decoding from Sqids-encoded strings via encoding.TextUnmarshaler.
// Use it in URI/query binding structs with ShouldBindUri
// or ShouldBindQuery for automatic Sqid decoding.
type SqID int64

// Int64 returns the underlying int64 value.
func (s SqID) Int64() int64 {
	return int64(s)
}

// UnmarshalText implements encoding.TextUnmarshaler,
// allowing Gin 1.12+ to automatically decode Sqids
// in URI and query parameter binding.
func (s *SqID) UnmarshalText(text []byte) error {
	if globalEncoder == nil {
		return fmt.Errorf(
			"global ID encoder not initialized",
		)
	}
	decoded, err := globalEncoder.Decode(string(text))
	if err != nil {
		return fmt.Errorf("invalid encoded id")
	}
	*s = SqID(decoded)
	return nil
}

// NewSqidsEncoder constructs an IDEncoder backed by the Sqids algorithm,
// using the given alphabet and minimum encoded string length.
func NewSqidsEncoder(
	alphabet string,
	minLength int,
) IDEncoder {
	if minLength < 0 || minLength > math.MaxUint8 {
		panic(fmt.Errorf(
			"invalid sqids minLength %d",
			minLength,
		))
	}

	s, err := sqids.New(sqids.Options{
		Alphabet: alphabet,
		MinLength: uint8(
			minLength,
		),
		// positive integer for padding, not security-sensitive.
	})
	if err != nil {
		panic(fmt.Errorf("failed to create sqids encoder: %w", err))
	}

	return &sqidsEncoder{sqids: s}
}

func (s *sqidsEncoder) Encode(id int64) string {
	if id < 0 {
		panic(fmt.Errorf("invalid negative id %d", id))
	}

	encoded, err := s.sqids.Encode(
		[]uint64{
			uint64(id),
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to encode id %d: %w", id, err))
	}
	return encoded
}

func (s *sqidsEncoder) Decode(
	encoded string,
) (int64, error) {
	decoded := s.sqids.Decode(encoded)
	if len(decoded) == 0 {
		return 0, fmt.Errorf("invalid encoded id")
	}
	if decoded[0] > math.MaxInt64 {
		return 0, fmt.Errorf("invalid encoded id")
	}
	//nolint:gosec // decoded[0] is range-checked against math.MaxInt64 above
	return int64(decoded[0]), nil
}
