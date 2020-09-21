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

// args contain single value - binary encored RangeRes
func (op IAVLRangeOp) Run(args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("value size is not 1")
	}

	valueBytes := args[0]
	var value RangeRes
	cdc.MustUnmarshalBinaryLengthPrefixed(valueBytes, &value)

	if len(value.Keys) != len(value.Values) {
		return nil, errors.New("keys length doesn't match values length")
	}

	// Compute the root hash and assume it is valid.
	// The caller checks the ultimate root later.
	root := op.Proof.ComputeRootHash()
	err := op.Proof.Verify(root)
	if err != nil {
		return nil, errors.Wrap(err, "computing root hash")
	}

	// TODO: Think more about validation logic. Need to check sequence and borders.
	for i, _ := range value.Keys {
		err := op.Proof.VerifyItem(value.Keys[i], value.Values[i])
		if err != nil {
			return nil, errors.Wrap(err, "verifying value")
		}
	}

	return [][]byte{root}, nil
}

func (op IAVLRangeOp) GetKey() []byte {
	return op.key
}
