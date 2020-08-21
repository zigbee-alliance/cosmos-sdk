package iavl

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/crypto/merkle"
)

const ProofOpIAVLRange = "iavl:r"

// ! Should be in tendermint/iavl.
//
// Similar to `github.com/tendermint/iavl@v0.12.4/proof_iavl_absence.go`
// and `github.com/tendermint/iavl@v0.12.4/proof_iavl_value.go`, but for range.
type IAVLRangeOp struct {
	// Binary encoded RangeReq.
	key []byte
	// Inner tree proof.
	Proof *iavl.RangeProof `json:"proof"`
}

var _ merkle.ProofOperator = IAVLRangeOp{}

func NewIAVLRangeOp(key []byte, proof *iavl.RangeProof) IAVLRangeOp {
	return IAVLRangeOp{
		key:   key,
		Proof: proof,
	}
}

func IAVLRangeOpDecoder(pop merkle.ProofOp) (merkle.ProofOperator, error) {
	if pop.Type != ProofOpIAVLRange {
		return nil, errors.Errorf("unexpected ProofOp.Type; got %v, want %v", pop.Type, ProofOpIAVLRange)
	}

	var op IAVLRangeOp
	err := cdc.UnmarshalBinaryLengthPrefixed(pop.Data, &op)
	if err != nil {
		return nil, errors.Wrap(err, "decoding ProofOp.Data into IAVLValueOp")
	}

	op.key = pop.Key

	return op, nil
}

func (op IAVLRangeOp) ProofOp() merkle.ProofOp {
	bz := cdc.MustMarshalBinaryLengthPrefixed(op)
	return merkle.ProofOp{
		Type: ProofOpIAVLRange,
		Key:  op.key,
		Data: bz,
	}
}

func (op IAVLRangeOp) String() string {
	return fmt.Sprintf("IAVLRangeOp{%v}", op.GetKey())
}

func (op IAVLRangeOp) Run(args [][]byte) ([][]byte, error) {
	// TODO
	//if len(args) != 1 {
	//	return nil, errors.New("Value size is not 1")
	//}
	//value := args[0]

	// Compute the root hash and assume it is valid.
	// The caller checks the ultimate root later.
	root := op.Proof.ComputeRootHash()
	err := op.Proof.Verify(root)
	if err != nil {
		return nil, errors.Wrap(err, "computing root hash")
	}

	// XXX What is the encoding for keys?
	// We should decode the key depending on whether it's a string or hex,
	// maybe based on quotes and 0x prefix?
	//err = op.Proof.VerifyItem([]byte(op.key), value)
	//if err != nil {
	//	return nil, errors.Wrap(err, "verifying value")
	//}

	return [][]byte{root}, nil
}

func (op IAVLRangeOp) GetKey() []byte {
	return op.key
}
