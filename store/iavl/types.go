package iavl

type RangeReq struct {
	StartKey []byte
	EndKey   []byte
	Limit    int
}

type RangeRes struct {
	keys   [][]byte
	values [][]byte
}
