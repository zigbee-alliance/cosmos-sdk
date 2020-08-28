package iavl

type RangeReq struct {
	StartKey []byte
	EndKey   []byte
	Limit    int
}

type RangeRes struct {
	Keys   [][]byte
	Values [][]byte
}
