module github.com/mctofu/musiclib

go 1.13

require (
	github.com/dhowden/tag v0.0.0-20200412032933-5d76b8eaae27
	github.com/mctofu/music-library-grpc v0.0.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sys v0.0.0-20200409092240-59c9f1ba88fa // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200410110633-0848e9f44c36 // indirect
	google.golang.org/grpc v1.29.1
)

replace github.com/mctofu/music-library-grpc => ../music-library-grpc
