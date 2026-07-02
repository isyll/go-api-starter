package idenc

import (
	"fmt"
	"math"
	"sync"

	"github.com/sqids/sqids-go"
)

type IDEncoder interface {
	Encode(id int64) string
	Decode(encoded string) (int64, error)
}

type sqidsEncoder struct {
	sqids *sqids.Sqids
}

var (
	globalEncoder IDEncoder
	encoderOnce   sync.Once
)

func SetGlobalEncoder(enc IDEncoder) {
	encoderOnce.Do(func() {
		globalEncoder = enc
	})
}

type SqID int64

func (s SqID) Int64() int64 {
	return int64(s)
}

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
