module github.com/lawrsp/pigo

go 1.18

require (
	github.com/lawrsp/pigo/generator v0.0.0-00010101000000-000000000000
	github.com/lawrsp/stringstyles v1.0.0
	github.com/urfave/cli v1.22.9
	golang.org/x/tools v0.1.12
)

replace github.com/lawrsp/pigo/generator => ./generator

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
