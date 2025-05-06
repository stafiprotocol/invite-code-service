package utils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

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
