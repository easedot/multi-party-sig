package btc

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
	"log"
	"math/big"
)

type ecdsa struct {
	R, S *big.Int
}

func IsOddBit(x []byte) uint32 {
	return uint32(x[0] & 1)
}
func CompressKey(x, y []byte) ([]byte, error) {
	out := make([]byte, 33)
	// we clone v to not case a race during a hash.Write
	// Doing it this way is compatible with Bitcoin
	if y[0] == 1 {

	}
	out[0] = byte(IsOddBit(y)) + 2
	data := x
	copy(out[1:], data[:])
	return out, nil
}

func Address(x, y []byte) string {
	var prefix []byte
	//for production
	prefix, _ = hex.DecodeString("00")
	//for test
	//prefix, _ = hex.DecodeString("6F")
	compressAddress, err := CompressKey(x, y)
	if err != nil {
		log.Printf("error:%s", err)
	}
	h160 := append(prefix, hash160(compressAddress)...)
	chksum := dblSha256(h160)
	return base58.Encode(append(h160, chksum[:4]...))
}

func Signature(r, s []byte) []byte {
	encoding := ByteToDER(r, s)
	// Append sighashtype to the signature (required) (01 = ALL)
	encoding = append(encoding, []byte{0x01}...)
	return encoding
}

func SigCheck(sig []byte) []byte {
	//ec.ParseSignature()

	sigc, err := ec.ParseDERSignature(sig)
	if err != nil {
		log.Printf("error:%s", err)
	}
	log.Printf("CheckDERSig:%x", sigc.Serialize())
	return sigc.Serialize()
}

func ByteToDER(rb, sb []byte) []byte {
	type ecdsa struct {
		R, S *big.Int
	}
	r := new(big.Int).SetBytes(rb)
	s := new(big.Int).SetBytes(sb)

	sequence := ecdsa{r, s}
	encoding, _ := asn1.Marshal(sequence)

	return encoding
}

//func PointsToDER(r, s *big.Int) []byte {
//	return BytesToDER(r.Bytes(), s.Bytes())
//}

//// PointsToDER Convert an ECDSA signature (points R and S) to a byte array using ASN.1 DER encoding.
//// This is a port of Bitcore's Key.rs2DER method.
//func BytesToDER(r, s []byte) []byte {
//	// Ensure MSB doesn't break big endian encoding in DER sigs
//	prefixPoint := func(b []byte) []byte {
//		if len(b) == 0 {
//			b = []byte{0x00}
//		}
//		if b[0]&0x80 != 0 {
//			paddedBytes := make([]byte, len(b)+1)
//			copy(paddedBytes[1:], b)
//			b = paddedBytes
//		}
//		return b
//	}
//
//	rb := prefixPoint(r)
//	sb := prefixPoint(s)
//	// DER encoding:
//	// 0x30 + z + 0x02 + len(rb) + rb + 0x02 + len(sb) + sb
//	length := 2 + len(rb) + 2 + len(sb)
//
//	der := append([]byte{0x30, byte(length), 0x02, byte(len(rb))}, rb...)
//	der = append(der, 0x02, byte(len(sb)))
//	der = append(der, sb...)
//
//	encoded := make([]byte, hex.EncodedLen(len(der)))
//	hex.Encode(encoded, der)
//	log.Printf("pe:%x", encoded)
//	return encoded
//}
//
//// Get the X and Y points from a DER encoded signature
//// Sometimes demarshalling using Golang's DEC to struct unmarshalling fails; this extracts R and S from the bytes
//// manually to prevent crashing.
//// This should NOT be a hex encoded byte array
//func PointsFromDER(der []byte) (R, S *big.Int) {
//	R, S = &big.Int{}, &big.Int{}
//
//	data := asn1.RawValue{}
//	if _, err := asn1.Unmarshal(der, &data); err != nil {
//		panic(err.Error())
//	}
//
//	// The format of our DER string is 0x02 + rlen + r + 0x02 + slen + s
//	rLen := data.Bytes[1] // The entire length of R + offset of 2 for 0x02 and rlen
//	r := data.Bytes[2 : rLen+2]
//	// Ignore the next 0x02 and slen bytes and just take the start of S to the end of the byte array
//	s := data.Bytes[rLen+4:]
//
//	R.SetBytes(r)
//	S.SetBytes(s)
//
//	return
//}

func hash160(data []byte) []byte {
	sha := sha256.New()
	ripe := ripemd160.New()
	sha.Write(data)
	ripe.Write(sha.Sum(nil))
	return ripe.Sum(nil)
}

func dblSha256(data []byte) []byte {
	sha1 := sha256.New()
	sha2 := sha256.New()
	sha1.Write(data)
	sha2.Write(sha1.Sum(nil))
	return sha2.Sum(nil)
}
