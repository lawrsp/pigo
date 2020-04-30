package builder

import (
	"log"
	"reflect"
	"strings"

	"github.com/lawrsp/pigo/pkg/parser"
)

type Field struct {
	Name  string
	Field *parser.Field
}

type FieldList struct {
	TagName string
	Fields  []*Field
}

func NewFieldList(tagName string) *FieldList {
	return &FieldList{
		TagName: tagName,
		Fields:  []*Field{},
	}
}

func (list *FieldList) Print() {
	for i, fd := range list.Fields {
		log.Printf("%d,%s: %s(%s)", i, fd.Name, fd.Field.Name(), fd.Field.Type)
	}
}
func (list *FieldList) GetFieldByName(name string) *Field {
	for _, fd := range list.Fields {
		if fd.Name == name {
			return fd
		}
	}
	return nil
}
func (list *FieldList) SpreadInspector(field *parser.Field) bool {
	name := ""
	if list.TagName != "" {
		stag := reflect.StructTag(field.Tag).Get(list.TagName)
		if stag != "" {
			vtags := strings.Split(stag, ",")
			if vtags[0] == "-" {
				return false
			}
			name = vtags[0]
		}
	}

	if name == "" {
		if field.IsAnonymous() && field.IsStruct() {
			//anonymous struct without specified name, step into
			return true
		}
		//has name
		name = field.Name()
		// name = stringutils.SnakeCase(name)
	}
	list.Fields = append(list.Fields, &Field{
		Name:  name,
		Field: field,
	})

	return false
}

func (list *FieldList) ContainPathToType(t parser.Type) (*Field, []*parser.TPath, bool) {
	for _, fd := range list.Fields {
		if paths, ok := parser.TypeToType(fd.Field.Type, t, []*parser.TPath{}); ok {
			return fd, paths, true
		}
	}
	return nil, nil, false
}

func GetStructHolder(t parser.Type) parser.Type {
	// *Struct / Struct
	typeHolder := t.Copy()
	if holder, ok := typeHolder.Underlying().(*parser.PointerType); ok {
		if holder.Stars > 1 {
			holder.Stars = 1
		}
	}

	return typeHolder
}
