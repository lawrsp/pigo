package genrpc

// For example, given this common services:
// ```go
//  package services
//
//  func (*x)GetUserById(db Database, id uint64) (*User, error) {
//     ....
//  }
//  func (*x) GetUserListByIds(db Database, ids []uint64) ([]User, error) {
//
//  }
//
// ```
// ```go
//  package pb
//
//  type UserManagerServer interface {
//      GetId(ctx context.Context, *IdRequest) error
//      GetIds(ctx context.Context, *IdListRequest) error
//  }
// ```
//
// Write a yaml config x.yaml:
//
// ```yaml
//  version: "1"
//  files: "user.go"
//  imports:
//      userpb: "github.com/lawrsp/gomms/protos/user"
//      user: "github.com/lawrsp/gomms/mdoels/user"
//      util: "github.com/lawrsp/gomms/rpc/utils"
//  output: "rpc_gen.go"
//  sender: "UserRpc"
//  interface: "userpb.UserManagerServer"
//  functions:
//      util_error: utils.SetError
//      util_db: utils.GetDatabase
//
//  gnerates:
//      GetId:
//         name: GetId
//         pre_functions:
//              - util_db
//         error_function: util_error
//         service:
//             sender: user.UserService
//             function: GetUserById
//         assign:
//             source: user.User
//             target: userpb.User
//             convert: UserProto
//      GetIds:
//         pre_functions:
//              - util_db
//         error_function: util_error
//         service:
//              sender: user.UserSerivce
//              function: GetUserListByIds
//         assign:
//              source: user.User
//              target: userpb.User
//              convert: UserProto
//              ext: array
// ```
//
// then generate:
// func (s *UserRpc)GetId(ctx context.Context, req *IdRequrest) error {
//  .....
//}
