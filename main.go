package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tranvictor/ethashproof/ethash"
	"github.com/tranvictor/ethashproof/mtree"
	"github.com/tranvictor/ethutils/reader"
)

func processDuringRead(
	datasetPath string, mt *mtree.DagTree) {
	var f *os.File
	var err error
	for {
		f, err = os.Open(datasetPath)
		if err == nil {
			break
		} else {
			fmt.Printf("Reading DAG file %s failed with %s. Retry in 10s...\n", datasetPath, err.Error())
			time.Sleep(10 * time.Second)
		}
	}
	r := bufio.NewReader(f)
	buf := [128]byte{}
	// ignore first 8 bytes magic number at the beginning
	// of dataset. See more at https://gopkg.in/ethereum/wiki/wiki/Ethash-DAG-Disk-Storage-Format
	_, err = io.ReadFull(r, buf[:8])
	if err != nil {
		log.Fatal(err)
	}
	var i uint32 = 0
	for {
		n, err := io.ReadFull(r, buf[:128])
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		if n != 128 {
			log.Fatal("Malformed dataset")
		}
		mt.Insert(mtree.Word(buf), i)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		i++
	}
}

func main() {
	r := reader.NewEthReader()
	header, err := r.HeaderByNumber(2182207)
	if err != nil {
		fmt.Printf("Getting header failed: %s\n", err)
		return
	}

	blockno := header.Number.Uint64()
	indices := ethash.Instance.GetVerificationIndices(
		blockno,
		HashHeaderNoNonce(header),
		header.Nonce.Uint64(),
	)

	dt := mtree.NewDagTree()
	dt.RegisterIndex(indices...)

	ethash.MakeDAG(blockno, ethash.DefaultDir)

	fullSize := ethash.DAGSize(blockno)
	fullSizeIn128Resolution := fullSize / 128
	branchDepth := len(fmt.Sprintf("%b", fullSizeIn128Resolution-1))
	dt.RegisterStoredLevel(uint32(branchDepth), uint32(10))

	path := ethash.PathToDAG(uint64(blockno/30000), ethash.DefaultDir)
	fmt.Printf("Calculating the proofs...\n")
	start := time.Now()
	processDuringRead(path, dt)

	dt.Finalize()

	elements := []*big.Int{}
	for _, w := range dt.AllDAGElements() {
		elements = append(elements, w.ToUint256Array()...)
	}
	fmt.Printf("DAG elements: %v\n", elements)

	allProofs := []*big.Int{}
	for _, be := range dt.AllBranchesArray() {
		allProofs = append(allProofs, be.Big())
	}
	fmt.Printf("DAG element proofs: %v\n", allProofs)
	end := time.Now()
	fmt.Printf("Proof calculation took: %s\n", common.PrettyDuration(end.Sub(start)))
	fmt.Printf("DAG merkle root: %s\n", dt.RootHash().Hex())
}
