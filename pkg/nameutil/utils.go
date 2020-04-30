package nameutil

import (
	"fmt"
)

func IncreaseName(name string) string {
	len := len(name)
	suffix := make([]byte, len)
	nameBytes := []byte(name)

	end := len - 1
	for ; end > 0; end -= 1 {
		if nameBytes[end] >= '0' && nameBytes[end] <= '9' {
			suffix[end] = nameBytes[end]
		} else {
			break
		}
	}

	num := 0
	for idx := end + 1; idx < len; idx += 1 {
		num = num*10 + int((suffix[idx] - '0'))
	}
	num += 1

	return fmt.Sprintf("%s%d", nameBytes[:end+1], num)
}
