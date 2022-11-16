module socket.io

go 1.16

retract v1.0.8

require (
	engine.io v0.0.0-00010101000000-000000000000
	github.com/andybalholm/brotli v1.0.4
	github.com/mitchellh/mapstructure v1.5.0
)

replace engine.io => github.com/edelbrocken/engine.io v0.0.0-20221116102827-d7869cbc1972
