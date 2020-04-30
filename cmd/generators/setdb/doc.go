package setdb

// For example, given this snippet,
//
//  package painkiller
//
//  type Pill struct {
//      Id *uint64
//      Name *string
//  }
//
//
// running this command
//
//  setdb -type=Pill
//
// in the same directory will create the file pill_setdb.go, in package painkiller,
// containing a definition of
//
//  func (p *Pill) SetDB(db *orm.SQLDB) *orm.SQLDB
//
//
// Typically this process would be run using go generate, like this:
//
//  //go:generate setdb -type=Pill
//
