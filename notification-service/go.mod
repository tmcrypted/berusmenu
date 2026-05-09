module github.com/tmcrypted/berusmenu/notification-service

go 1.25.1

replace github.com/tmcrypted/berusmenu/proto => ../proto

require (
	github.com/jackc/pgx/v5 v5.9.2
	github.com/segmentio/kafka-go v0.4.51
	github.com/tmcrypted/berusmenu/proto v0.0.0-00010101000000-000000000000
	go.uber.org/mock v0.6.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.33.0 // indirect
)
