package builder

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/lawrsp/pigo/generator/parser"
)

//camel string, xx_yy to xxYy
func CamelString(src string) string {
	s := []byte(src)
	data := make([]byte, 0, len(s))
	j := false
	k := true
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if i == 0 && d >= 'A' && d <= 'Z' {
			d = d + 32
		}
		if !k && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || !k) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

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

var definedNames = map[string]string{
	"context":   "ctx",
	"Context":   "Ctx",
	"error":     "err",
	"Error":     "Err",
	"Server":    "Srv",
	"server":    "srv",
	"Interface": "Itf",
	"interface": "itf",
	"int":       "i",
	"uint":      "ui",
}

func checkDefinedName(name string) (result string, pos int) {
	for k, v := range definedNames {
		if strings.HasPrefix(name, k) {
			return v, len(k)
		}
	}

	return "", 0
}

func genVariableName(name string) (result string, isList bool) {

	nn := make([]byte, len(name))
	i := 0
	ignoreBrackets := false

	for pos := 0; pos < len(name); {
		b := name[pos]

		if (b >= 'A' && b <= 'Z') || i == 0 {
			defined, nextPos := checkDefinedName(name[pos:])
			if nextPos > 0 {
				result = string(nn[:i]) + defined
				pos += nextPos
				i = 0

				continue
			}
		}

		//ignore [....]
		if b == '[' {
			ignoreBrackets = true
		}
		if ignoreBrackets {
			pos++
			if b == ']' {
				ignoreBrackets = false
				isList = true
			}
			continue
		}

		switch b {
		case '.':
			//dot seperator, ignore ealiar
			result = ""
			i = 0
		case 'a', 'e', 'i', 'o', 'u':
			//, 'A', 'E', 'I', 'O', 'U':
			//ignore vowels, except the first one
			if i == 0 {
				nn[i] = b
				i++
			}
		case '*':
			//ignore pointer
		default:
			nn[i] = b
			i++
		}
		pos++
	}

	result = result + string(nn[:i])
	return
}

func VariableName(name string) string {
	result, isList := genVariableName(name)
	if isList {
		result = result + "List"
	}

	return CamelString(result)
}

func GenNameFromType(typ ast.Expr) string {
	t := parser.ParseType(typ)
	name := VariableName(t.String())
	return name
}
