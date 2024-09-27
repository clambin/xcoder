package bits

import "strconv"

type Bits int

func (b Bits) Format(decimals int) string {
	floatBits := float64(b)
	unit := "b"
	if floatBits > 1000 {
		floatBits /= 1000
		unit = "kb"
	}
	if floatBits > 1000 {
		floatBits /= 1000
		unit = "mb"
	}
	return strconv.FormatFloat(floatBits, 'f', decimals, 64) + " " + unit
}
