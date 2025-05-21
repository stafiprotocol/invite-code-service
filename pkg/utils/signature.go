package utils

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

// Signature verification follows Ethereum's personal_sign standard.
//
// 👉 Frontend should construct the message exactly as below for signing:
//
//
// Example bind message to be signed(/api/v1/invite/bind):
//
// Please sign this message to verify your identity.
// This request will not trigger any blockchain transaction or cost any gas.
//
// Invite Code: abc123
// Discord ID: 987654321
// Discord Name: xxx
// Timestamp: 123456
//
//
// Example gen message to be signed(/api/v1/invite/genInviteCode):
//
// Please sign this message to verify your identity.
// This request will not trigger any blockchain transaction or cost any gas.
//
// Timestamp: 123456
//
//
// The full message is then signed using personal_sign (EIP-191).
//
// On backend, the message is prefixed with the standard:
//   "\x19Ethereum Signed Message:\n" + len(message) + message
// before recovering the signer address using the signature.
//
// ⚠️ Important:
// - The message must match exactly, including line breaks and spaces.
// - Timestamp is recommended to prevent replay attacks (±5 min validity).

func VerifySigsEthPersonal(sigs []byte, message string, address common.Address) bool {
	useMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	useSigs := make([]byte, 65)
	copy(useSigs, sigs)
	if useSigs[64] > 26 {
		useSigs[64] = useSigs[64] - 27
	}
	pubkey, err := ethCrypto.Ecrecover(ethCrypto.Keccak256([]byte(useMessage)), useSigs)
	if err != nil {
		return false
	}
	recoverAddress := common.BytesToAddress(ethCrypto.Keccak256(pubkey[1:])[12:])
	return recoverAddress == address
}

func IsValidSignTime(timestamp uint64) bool {
	if time.Now().Unix() > int64(timestamp+300) || time.Now().Unix()+300 < int64(timestamp) {
		return false
	}
	return true
}

func BuildBindMessage(inviteCode string, discordID, discordName string, timestamp uint64) string {
	return fmt.Sprintf(`Please sign this message to verify your identity.
This request will not trigger any blockchain transaction or cost any gas.

Invite Code: %s
Discord ID: %s
Discord Name: %s
Timestamp: %d`, inviteCode, discordID, discordName, timestamp)
}

func BuildGenMessage(timestamp uint64) string {
	return fmt.Sprintf(`Please sign this message to verify your identity.
This request will not trigger any blockchain transaction or cost any gas.

Timestamp: %d`, timestamp)
}
