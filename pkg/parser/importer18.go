// +build !go1.9

package parser

import (
	"go/importer"
	"go/types"
)

func defaultImporter() types.Importer {
	return importer.Default()
}
