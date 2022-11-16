module socket.io

go 1.16

retract v1.0.8

require (
	github.com/andybalholm/brotli v1.0.4
	github.com/mitchellh/mapstructure v1.5.0
)

require (
	engine.io v0.0.0
)

replace engine.io => ../engine.io
