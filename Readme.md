# Ethash proof - Command line to calculate ethash (ethereum POW) merkle proof

Ethashproof is a commandline to calculate proof data for an ethash POW, it is used by project `SmartPool` and a decentralized
bridge between Etherum and EOS developed by Kyber Network team.

## Features

1. Calculate merkle root of the ethash dag dataset with given epoch
2. Calculate merkle proof of the pow (dataset elements and their merkle proofs) given the pow submission with given block header
3. Generate dag dataset

## Merkle tree in ethashproof

In `ethashproof`, we construct a merkle tree out of ethash DAG dataset in order to get merkle root
of the dataset and get merkle proof for any specific dataset element.
Each DAG dataset is a sequence of many 128 bytes dataset elements, denoted as:
```
e0, e1, e2, ..., en
```

### Merkle tree explanation

The merkle tree is constructed in the following way:

- Step 1: Calculate hash of each element using `elementhash` function. We would have:
```
h00, h01, h02, h03, ..., h0n
 |    |    |    |   ...   |
e0,  e1,  e2,  e3,  ..., en
```
where `h01` means hash at level 0, element 1. The hash is 32 bytes.

- Step 2: In this step, set the `working level` to 0.
calculate `h10 = hash(h00, h01)`, `h11 = hash(h02, h03)`, ...
if at the end of the level, there is only one element left, we duplicate it and calculate the hash out of them, eg. `h0x = hash(h0x*2, h0x*2)`
```
   h10       h11    ...  h1n/2
  /  \      /  \           /  \
h00, h01, h02, h03, ..., h0n  h0n
 |    |    |    |   ...   |
e0,  e1,  e2,  e3,  ..., en
```

- Step 3: Increase `working level` by 1 and go back to Step 2 until there is only 1 element in the working level. That element is the merkle root.
```
            merkle root
                .
              .....
            ..........
          ...............
        h20         ...  h2n/4
      /     \          /   |
   h10       h11    ...  h1n/2
  /  \      /  \           /  \
h00, h01, h02, h03, ..., h0n  h0n
 |    |    |    |   ...   |
e0,  e1,  e2,  e3,  ..., en
```

### Hash function
0. Given keccak256()
1. Hash function for data element(`elementhash`)
`elementhash` returns 16 bytes hash of the dataset element.
```
function elementhash(data) => 16bytes {
  h = keccak256(conventional(data)) // conventional function is defined in dataset element encoding section
  return last16Bytes(h)
}
```

2. Hash function for 2 sibling nodes (`hash`)
`hash` returns 16 bytes hash of 2 consecutive elements in a working level.
```
function hash(a, b) => 16bytes {
  h = keccak256(zeropadded(a), zeropadded(b)) // where zeropadded function prepend 16 bytes of 0 to its param
  return last16Bytes(h)
}
```

### Dataset element encoding
In order to make it easy (and gas saving) for ethereum smart contract (the earliest contract we used to verify the proof) to work with the
dataset element, `ethashproof` use a conventional encoding for the dataset element as defined below:

1. assume the element is `abcd` where a, b, c, d are 32 bytes word
2. `first = concat(reverse(a), reverse(b))` where `reverse` reverses the bytes
3. `second = concat(reverse(c), reverse(d))`
4. conventional encoding of `abcd` is `concat(first, second)`

### Merkle proof branch
1. Explanation
Please read more on http://www.certificate-transparency.org/log-proofs-work at Merkle Audit Proofs section.

2. Merkle proof element encoding

