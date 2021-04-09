// Package ecdsa provides a crypto implementation for HotStuff using Go's 'crypto/ecdsa' package.
package ecdsa

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"

	"github.com/relab/hotstuff"
	"github.com/relab/hotstuff/crypto"
	"go.uber.org/multierr"
)

const (
	// PrivateKeyFileType is the PEM type for a private key.
	PrivateKeyFileType = "ECDSA PRIVATE KEY"

	// PublicKeyFileType is the PEM type for a public key.
	PublicKeyFileType = "ECDSA PUBLIC KEY"
)

// Signature is an ECDSA signature
type Signature struct {
	r, s   *big.Int
	signer hotstuff.ID
}

// RestoreSignature restores an existing signature. It should not be used to create new signatures, use Sign instead.
func RestoreSignature(r, s *big.Int, signer hotstuff.ID) *Signature {
	return &Signature{r, s, signer}
}

// Signer returns the ID of the replica that generated the signature.
func (sig Signature) Signer() hotstuff.ID {
	return sig.signer
}

// R returns the r value of the signature
func (sig Signature) R() *big.Int {
	return sig.r
}

// S returns the s value of the signature
func (sig Signature) S() *big.Int {
	return sig.s
}

// ToBytes returns a raw byte string representation of the signature
func (sig Signature) ToBytes() []byte {
	var b []byte
	b = append(b, sig.r.Bytes()...)
	b = append(b, sig.s.Bytes()...)
	return b
}

var _ hotstuff.Signature = (*Signature)(nil)

// ThresholdSignature is a set of (partial) signatures that form a valid threshold signature when there are a quorum
// of valid (partial) signatures.
type ThresholdSignature map[hotstuff.ID]*Signature

// RestoreThresholdSignature should only be used to restore an existing threshold signature from a set of signatures.
// To create a new verifiable threshold signature, use CreateThresholdSignature instead.
func RestoreThresholdSignature(signatures []*Signature) ThresholdSignature {
	sig := make(ThresholdSignature, len(signatures))
	for _, s := range signatures {
		sig[s.signer] = s
	}
	return sig
}

// ToBytes returns the object as bytes.
func (sig ThresholdSignature) ToBytes() []byte {
	var b []byte
	// sort by ID to make it deterministic
	order := make([]hotstuff.ID, 0, len(sig))
	for _, signature := range sig {
		i := sort.Search(len(order), func(i int) bool { return signature.signer < order[i] })
		order = append(order, 0)
		copy(order[i+1:], order[i:])
		order[i] = signature.signer
	}
	for _, id := range order {
		b = append(b, sig[id].ToBytes()...)
	}
	return b
}

// Participants returns the IDs of replicas who participated in the threshold signature.
func (sig ThresholdSignature) Participants() hotstuff.IDSet {
	return sig
}

// Add adds an ID to the set.
func (sig ThresholdSignature) Add(id hotstuff.ID) {
	panic("not implemented")
}

// Contains returns true if the set contains the ID.
func (sig ThresholdSignature) Contains(id hotstuff.ID) bool {
	_, ok := sig[id]
	return ok
}

// ForEach calls f for each ID in the set.
func (sig ThresholdSignature) ForEach(f func(hotstuff.ID)) {
	for id := range sig {
		f(id)
	}
}

var _ hotstuff.ThresholdSignature = (*ThresholdSignature)(nil)
var _ hotstuff.IDSet = (*ThresholdSignature)(nil)

type ecdsaCrypto struct {
	mod *hotstuff.HotStuff
}

func (ec *ecdsaCrypto) InitModule(hs *hotstuff.HotStuff) {
	ec.mod = hs
}

// New returns a new signer and a new verifier.
func New() hotstuff.CryptoImpl {
	ec := &ecdsaCrypto{}
	return ec
}

func (ec *ecdsaCrypto) getPrivateKey() *ecdsa.PrivateKey {
	pk := ec.mod.PrivateKey()
	return pk.(*ecdsa.PrivateKey)
}

// Sign signs a hash.
func (ec *ecdsaCrypto) Sign(hash hotstuff.Hash) (sig hotstuff.Signature, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, ec.getPrivateKey(), hash[:])
	if err != nil {
		return nil, err
	}
	return &Signature{
		r:      r,
		s:      s,
		signer: ec.mod.ID(),
	}, nil
}

// Verify verifies a signature given a hash.
func (ec *ecdsaCrypto) Verify(sig hotstuff.Signature, hash hotstuff.Hash) bool {
	_sig, ok := sig.(*Signature)
	if !ok {
		return false
	}
	replica, ok := ec.mod.Config().Replica(sig.Signer())
	if !ok {
		ec.mod.Logger().Infof("ecdsaCrypto: got signature from replica whose ID (%d) was not in the config.", sig.Signer())
		return false
	}
	pk := replica.PublicKey().(*ecdsa.PublicKey)
	return ecdsa.Verify(pk, hash[:], _sig.R(), _sig.S())
}

// CreateThresholdSignature creates a threshold signature from the given partial signatures.
func (ec *ecdsaCrypto) CreateThresholdSignature(partialSignatures []hotstuff.Signature, hash hotstuff.Hash) (_ hotstuff.ThresholdSignature, err error) {
	thrSig := make(ThresholdSignature)
	for _, s := range partialSignatures {
		if thrSig.Participants().Contains(s.Signer()) {
			err = multierr.Append(err, crypto.ErrPartialDuplicate)
			continue
		}

		sig, ok := s.(*Signature)
		if !ok {
			err = multierr.Append(err, fmt.Errorf("%w: %T", crypto.ErrWrongType, s))
			continue
		}

		// use the registered verifier instead of ourself to verify.
		// this makes it possible for the signatureCache to work.
		if ec.mod.Crypto().Verify(s, hash) {
			thrSig[sig.signer] = sig
		}
	}

	if len(thrSig) >= ec.mod.Config().QuorumSize() {
		return thrSig, nil
	}

	return nil, multierr.Combine(crypto.ErrNotAQuorum, err)
}

// VerifyThresholdSignature verifies a threshold signature.
func (ec *ecdsaCrypto) VerifyThresholdSignature(signature hotstuff.ThresholdSignature, hash hotstuff.Hash) bool {
	sig, ok := signature.(ThresholdSignature)
	if !ok {
		return false
	}
	if len(sig) < ec.mod.Config().QuorumSize() {
		return false
	}
	results := make(chan bool)
	for _, pSig := range sig {
		go func(sig *Signature) {
			results <- ec.mod.Crypto().Verify(sig, hash)
		}(pSig)
	}
	numVerified := 0
	for range sig {
		if <-results {
			numVerified++
		}
	}
	return numVerified >= ec.mod.Config().QuorumSize()
}

var _ hotstuff.CryptoImpl = (*ecdsaCrypto)(nil)
