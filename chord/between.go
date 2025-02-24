package chord

import "math/big"

func between(start, id, end *big.Int) bool {
	if start.Cmp(end) < 0 {
		return start.Cmp(id) < 0 && id.Cmp(end) < 0
	}
	return start.Cmp(id) < 0 || id.Cmp(end) < 0
}
