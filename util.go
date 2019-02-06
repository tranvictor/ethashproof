package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

func HashHeaderNoNonce(h *types.Header) common.Hash {
	return RLPHash([]interface{}{
		h.ParentHash,
		h.UncleHash,
		h.Coinbase,
		h.Root,
		h.TxHash,
		h.ReceiptHash,
		h.Bloom,
		h.Difficulty,
		h.Number,
		h.GasLimit,
		h.GasUsed,
		h.Time,
		h.Extra,
	})
}

func RLPHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak512()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
