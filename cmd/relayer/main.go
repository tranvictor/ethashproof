package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tranvictor/ethashproof"
	"github.com/tranvictor/ethashproof/ethash"
	"github.com/tranvictor/ethashproof/mtree"
	"github.com/tranvictor/ethutils/reader"
)

type Output struct {
	MerkleRoot   string   `json:"merkle_root"`
	Elements     []string `json:"elements"`
	MerkleProofs []string `json:"merkle_proofs"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Block number param is missing. Please run ./relayer <blocknumber> instead.\n")
		return
	}
	if len(os.Args) > 2 {
		fmt.Printf("Please pass only 1 param as a block number. Please run ./relayer <blocknumber> instead.\n")
		return
	}
	number, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Please pass a number as a block number. Please run ./relayer <integer> instead.\n")
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("Getting block header\n")
	r := reader.NewEthReader()
	header, err := r.HeaderByNumber(int64(number))
	if err != nil {
		fmt.Printf("Getting header failed: %s\n", err)
		return
	}

	blockno := header.Number.Uint64()
	epoch := blockno / 30000
	cache, err := ethashproof.LoadCache(int(epoch))
	if err != nil {
		fmt.Printf("Cache is missing, calculate dataset merkle tree to create the cache first...\n")
		_, err = ethashproof.CalculateDatasetMerkleRoot(epoch, true)
		if err != nil {
			fmt.Printf("Creating cache failed: %s\n", err)
			return
		}
		cache, err = ethashproof.LoadCache(int(epoch))
		if err != nil {
			fmt.Printf("Getting cache failed after trying to creat it: %s. Abort.\n", err)
			return
		}
	}

	indices := ethash.Instance.GetVerificationIndices(
		blockno,
		ethash.Instance.SealHash(header),
		header.Nonce.Uint64(),
	)

	fmt.Printf("Proof length: %d\n", cache.ProofLength)

	output := &Output{
		MerkleRoot:   cache.RootHash.Hex(),
		Elements:     []string{},
		MerkleProofs: []string{},
	}

	for _, index := range indices {
		element, proof, err := ethashproof.CalculateProof(blockno, index, cache)
		if err != nil {
			fmt.Printf("calculating the proofs failed for index: %d, error: %s\n", index, err)
			return
		}
		es := element.ToUint256Array()
		for _, e := range es {
			output.Elements = append(output.Elements, hexutil.EncodeBig(e))
		}
		allProofs := []*big.Int{}
		for _, be := range mtree.HashesToBranchesArray(proof) {
			allProofs = append(allProofs, be.Big())
		}
		for _, pr := range allProofs {
			output.MerkleProofs = append(output.MerkleProofs, hexutil.EncodeBig(pr))
		}
	}

	fmt.Printf("Json output:\n\n")
	outputJson, _ := json.Marshal(output)
	fmt.Printf("%s\n", outputJson)
}
