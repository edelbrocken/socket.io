module github.com/zishang520/socket.io

go 1.16

retract v1.0.8

require (
	github.com/andybalholm/brotli v1.0.4
	github.com/mitchellh/mapstructure v1.5.0
)

replace (
	github.com/zishang520/engine.io => github.com/edelbrocken/engine.io v0.0.0-20221116113430-92cf57d6b4a7
	github.com/zishang520/socket.io => github.com/edelbrocken/socket.io dev
)
