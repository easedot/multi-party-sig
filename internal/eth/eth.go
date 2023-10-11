package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
)

const (
	// compactSigSize is the size of a compact signature.  It consists of a
	// compact signature recovery code byte followed by the R and S components
	// serialized as 32-byte big-endian values. 1+32*2 = 65.
	// for the R and S components. 1+32+32=65.
	compactSigSize = 65

	// compactSigMagicOffset is a value used when creating the compact signature
	// recovery code inherited from Bitcoin and has no meaning, but has been
	// retained for compatibility.  For historical purposes, it was originally
	// picked to avoid a binary representation that would allow compact
	// signatures to be mistaken for other components.
	compactSigMagicOffset = 27

	// compactSigCompPubKey is a value used when creating the compact signature
	// recovery code to indicate the original public key was compressed.
	compactSigCompPubKey = 4

	// RecoveryIDOffset points to the byte offset within the signature that contains the recovery id.
	RecoveryIDOffset = 64
	// pubKeyRecoveryCodeOddnessBit specifies the bit that indicates the oddess
	// of the Y coordinate of the random point calculated when creating a
	// signature.
	pubKeyRecoveryCodeOddnessBit = 1 << 0

	// pubKeyRecoveryCodeOverflowBit specifies the bit that indicates the X
	// coordinate of the random point calculated when creating a signature was
	// >= N, where N is the order of the group.
	pubKeyRecoveryCodeOverflowBit = 1 << 1
)

func Address(x, y []byte) string {
	//for test
	//prefix, _ = hex.DecodeString("6F")
	hash := crypto.Keccak256(x, y)
	log.Printf("AD:%x", hash)
	addr := fmt.Sprintf("0x%x", hash[len(hash)-20:])
	return addr
}

func Sign(r, s []byte, v byte) ([]byte, error) {
	sig := sigCompact(r, s, v, false)
	// Convert to Ethereum signature format with 'recovery id' v at the end.
	ve := sig[0] - 27
	copy(sig, sig[1:])
	sig[RecoveryIDOffset] = ve
	return sig, nil
}

func sigCompact(r, s []byte, v byte, isCompressedKey bool) []byte {
	compactSigRecoveryCode := compactSigMagicOffset + v
	// Output <compactSigRecoveryCode><32-byte R><32-byte S>.
	if isCompressedKey {
		compactSigRecoveryCode += compactSigCompPubKey
	}

	var b [compactSigSize]byte
	b[0] = compactSigRecoveryCode
	copy(b[1:33], r)
	copy(b[33:65], s)
	return b[:]
}

func SigToPub(hash, sig []byte) ([]byte, error) {
	pub, err := crypto.Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

//func RecoverPubKey(sign, hash []byte) ([]byte, bool, error) {
//	pk, wasCompressed, err := ecdsa.RecoverCompact(sign, hash)
//	return pk.SerializeCompressed(), wasCompressed, err
//}
