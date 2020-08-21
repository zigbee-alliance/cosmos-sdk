package iavl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCdc(t *testing.T) {
	rangeQuery := RangeReq{
		StartKey: []byte{1, 2, 3},
		EndKey:   nil,
		Limit:    20,
	}

	bytes := cdc.MustMarshalBinaryBare(rangeQuery)

	var decoded RangeReq
	cdc.MustUnmarshalBinaryBare(bytes, &decoded)

	assert.NotNil(t, decoded)
	assert.Equal(t, rangeQuery.StartKey, decoded.StartKey)
	assert.Equal(t, rangeQuery.EndKey, decoded.EndKey)
	assert.Equal(t, rangeQuery.Limit, decoded.Limit)
}
