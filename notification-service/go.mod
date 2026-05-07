module github.com/tmcrypted/berusmenu/notification-service

go 1.25.1

replace github.com/tmcrypted/berusmenu/proto => ../proto

require (
	github.com/segmentio/kafka-go v0.4.51
	github.com/tmcrypted/berusmenu/proto v0.0.0-00010101000000-000000000000
	go.uber.org/mock v0.6.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)
