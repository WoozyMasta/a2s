package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

const challengeBucketDuration = time.Minute

const challengeSecretSize = 32

// ChallengeProvider issues and validates challenge tokens for one remote endpoint.
type ChallengeProvider interface {
	Issue(netip.AddrPort) a2s.Challenge
	Validate(netip.AddrPort, a2s.Challenge) bool
}

// statelessChallengeProvider creates challenge tokens from a secret,
// remote endpoint, and short time bucket.
// It stores no per-client state.
type statelessChallengeProvider struct {
	now    func() time.Time
	secret [challengeSecretSize]byte
}

// NewStatelessChallengeProvider creates a provider with a cryptographically random secret.
func NewStatelessChallengeProvider() (ChallengeProvider, error) {
	return newStatelessChallengeProvider(rand.Reader, time.Now)
}

// Issue creates a challenge bound to remote and the current time bucket.
func (p *statelessChallengeProvider) Issue(remote netip.AddrPort) a2s.Challenge {
	bucket := challengeBucket(p.now())
	return p.challenge(remote, bucket)
}

// Validate reports whether challenge belongs to remote
// and the current or immediately preceding time bucket.
func (p *statelessChallengeProvider) Validate(
	remote netip.AddrPort,
	challenge a2s.Challenge,
) bool {
	bucket := challengeBucket(p.now())
	for _, candidateBucket := range [2]int64{bucket, bucket - 1} {
		candidate := p.challenge(remote, candidateBucket)
		if hmac.Equal(challenge[:], candidate[:]) {
			return true
		}
	}

	return false
}

// newStatelessChallengeProvider creates a provider from an injectable secret source
// and clock for deterministic tests.
func newStatelessChallengeProvider(
	secretSource io.Reader,
	now func() time.Time,
) (*statelessChallengeProvider, error) {
	provider := &statelessChallengeProvider{now: now}
	if provider.now == nil {
		provider.now = time.Now
	}
	if _, err := io.ReadFull(secretSource, provider.secret[:]); err != nil {
		return nil, fmt.Errorf("%w: generate secret: %w", ErrChallengeProvider, err)
	}

	return provider, nil
}

// challenge derives one non-reserved token for remote and bucket.
func (p *statelessChallengeProvider) challenge(
	remote netip.AddrPort,
	bucket int64,
) a2s.Challenge {
	for counter := uint32(0); counter < 256; counter++ {
		mac := hmac.New(sha256.New, p.secret[:])
		mac.Write(challengeMessage(remote, bucket, counter))

		var challenge a2s.Challenge
		copy(challenge[:], mac.Sum(nil)[:len(challenge)])
		if subtle.ConstantTimeCompare(challenge[:], a2s.InitialChallenge[:]) != 1 {
			return challenge
		}
	}

	// Reaching this branch would require all 256 truncated MACs to equal the reserved token.
	// Keep the invariant explicit even in that impossible case.
	return a2s.Challenge{0, 0, 0, 0}
}

// challengeBucket maps time to the fixed-width validity window.
func challengeBucket(now time.Time) int64 {
	return now.Unix() / int64(challengeBucketDuration/time.Second)
}

// challengeMessage returns the canonical MAC input for one endpoint and bucket.
// IPv4-mapped IPv6 addresses are normalized to IPv4 first.
func challengeMessage(remote netip.AddrPort, bucket int64, counter uint32) []byte {
	addr := remote.Addr().Unmap()
	addressBytes, _ := addr.MarshalBinary()

	message := make([]byte, 0, 1+len(addressBytes)+2+8+4)
	// #nosec G115 -- netip addresses are at most 16 bytes long.
	message = append(message, byte(len(addressBytes)))
	message = append(message, addressBytes...)
	message = binary.BigEndian.AppendUint16(message, remote.Port())
	// #nosec G115 -- preserve the two's-complement representation of int64.
	message = binary.BigEndian.AppendUint64(message, uint64(bucket))
	message = binary.BigEndian.AppendUint32(message, counter)

	return message
}
