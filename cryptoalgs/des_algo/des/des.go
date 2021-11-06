package des

import (
    "io"
    "encoding/binary"
)

const (
    roundsAmount = 16
)

func EncryptData(key []byte, data io.Reader, output io.Writer) error {
    if len(key) > 7 {
        panic("Invalid length of key")
    }

    // Generate keys of each round
    keyUint := convertByteSliceToUint64(key)
    keys := generateKeys(keyUint)

    var block [8]byte
    var n int
    var err error

    for {
        if n, err = data.Read(block[:]); err == io.EOF {
            return nil
        } else if err != nil {
            return err
        }

        if n < 8 {
            for i := n; i < 8; i++ {
                block[i] = 0
            }
        }

        // Perform initial permutation
        blockUint := convertByteSliceToUint64(block[:])
        blockIPPermutated := performIPStraightPermutation(blockUint)

        var roundsResult [roundsAmount + 1][2]uint32
        roundsResult[0][0], roundsResult[0][1] = uint32(blockIPPermutated), uint32(blockIPPermutated >> 32)

        // Start rounds loop
        for round := 1; round <= roundsAmount; round++ {
            // L{i} = R{i - 1}
            roundsResult[round][0] = roundsResult[round - 1][1]

            // R{i} = L{i - 1} XOR f(R{i - 1}, k{i})
            roundsResult[round][1] = roundsResult[round - 1][1] ^ feistelFunction(roundsResult[round - 1][0], keys[round - 1])
        }

        var finalBlock uint64
        finalBlock = uint64(roundsResult[16][0]) + uint64(roundsResult[16][1]) << 32

        // Perform reverse of initial permutation
        resultBlock := performIPReversePermutation(finalBlock)

        if _, err := output.Write(convertUint64ToByteSlice(resultBlock)); err != nil {
            return err
        }
    }
}

func feistelFunction(block uint32, key uint64) (res uint32) {
    expandEFunctionPerformed := performEExpandFunction(block)
    expandEFunctionPerformed ^= key

    res = performSPermutation(expandEFunctionPerformed)
    res = performPPermutation(res)

    return
}

func generateKeys(key uint64) (res [16]uint64) {
    for i := 0; i < len(res); i++ {
        res[i] = key
    }

    return
}

func performIPStraightPermutation(block uint64) (res uint64) {
    for i := 0; i < len(IPPermutation); i++ {
        designatedBit := IPPermutation[i] - 1
        res |= block & (1 << designatedBit) >> designatedBit << i
    }

    return
}

func performEExpandFunction(block uint32) (res uint64) {
    for i := 0; i < len(ExpandEFunction); i++ {
        designatedBit := ExpandEFunction[i] - 1
        res |= uint64(block) & (1 << designatedBit) >> designatedBit << i
    }

    return res
}

func performSPermutation(block uint64) (res uint32) {
    for sPermutationBlock := 0; sPermutationBlock < 8; sPermutationBlock++ {
        res |= uint32(performSPermutationFor(sPermutationBlock, block)) << uint32(4 * sPermutationBlock)
    }

    return
}

func performSPermutationFor(sPermutationBlock int, block uint64) byte {
    block &= 0x3F << (sPermutationBlock * 6)
    block >>= sPermutationBlock * 6

    a := block & 1 + (block & 0x20 >> 0x10)
    b := (block & 0x1E) >> 1

    return SPermutation[sPermutationBlock][a][b]
}

func performPPermutation(block uint32) (res uint32) {
    for i := 0; i < len(PPermutation); i++ {
        res |= (block & (1 << (PPermutation[i] - 1))) << i
    }

    return res
}

func performIPReversePermutation(block uint64) (res uint64) {
    for i := 0; i < len(IPPermutation); i++ {
        designatedBit := IPPostPermutation[i] - 1
        res |= block & (1 << designatedBit) >> designatedBit << i
    }

    return
}

func convertByteSliceToUint64(block []byte) uint64 {
    if len(block) > 8 {
        panic("Invalid length of block")
    }

    if len(block) < 8 {
        for i := len(block); i < 8; i++ {
            block = append(block, 0)
        }
    }

    return binary.BigEndian.Uint64(block)
}

func convertUint64ToByteSlice(block uint64) []byte {
    res := make([]byte, 8)

    binary.BigEndian.PutUint64(res, block)
    return res
}
