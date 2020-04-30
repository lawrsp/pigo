package tagutil

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func GetTag(tag reflect.StructTag, tagname string) string {
	stag := tag.Get(tagname)
	return stag

}

func getTagValueSimple(tag reflect.StructTag, tagname string, pos int) string {
	stag := tag.Get(tagname)
	if len(stag) > 0 {
		stags := strings.Split(stag, ",")
		if len(stags) > pos {
			return stags[pos]
		}
		return ""
	}

	return ""
}

func getTagValueComposite(tag reflect.StructTag, tagname string, lvl2 string) string {
	stag := tag.Get(tagname)
	if len(stag) > 0 {
		stags := strings.Split(stag, ";")
		next := fmt.Sprintf("%s:", lvl2)
		for _, oneTag := range stags {
			if strings.HasPrefix(oneTag, next) {
				return oneTag[len(next):]
			}
		}
	}

	return ""
}

func GetTagValue(tag reflect.StructTag, find string) string {
	if find == "" {
		return ""
	}

	names := strings.Split(find, ".")
	if len(names) == 1 {
		return GetTag(tag, find)
	}

	tagname := names[0]
	next := names[1]

	re := regexp.MustCompile("/^[0-9]*$/")
	if re.Match([]byte(next)) {
		pos, _ := strconv.ParseInt(next, 10, 32)
		return getTagValueSimple(tag, tagname, int(pos))
	}
	return getTagValueComposite(tag, tagname, next)
}
