package convert

// For example, given this snippet,
// ```go
//  package animal
//
//  type Cat struct {
//      Id *uint64
//      Name *string
//  }
//  type Dog struct {
//      Id uint64
//      Name string
//  }
// ```
//
// Write a yaml config x.yaml:
//
// ```yaml
//  version: "1"
//  output: "pets.go"
//  gnerates:
//      cat_to_dog:
//         name: CatToDog
//         source: cat
//         target: dog
//  dir: "."
// ```
//
// then run
// ```
//    pigo convert --file x.yaml
// ```
//
// then it will generate a pet.go :
// ```go
// package animal
//
// func CatToDog(src *Cat, dst *Dog) error {
//    if src.Id != nil {
//      dst.Id = *src.Id
//    }
//    if src.Name != nil {
//      dst.Name = *src.Name
//    }
//    return nil
// }
//
// ```
//
// Typically this process would be run using go generate, like this:
//
//  example1:  //go:generate pigo convert --file x.yaml
