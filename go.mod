module github.com/blues/note-go

// 2023-02-26 Raspberry Pi apt-get only is updated to 1.15
go 1.15

require (
	github.com/gofrs/flock v0.7.1
	github.com/shirou/gopsutil/v3 v3.21.6
	github.com/stretchr/testify v1.7.0
	go.bug.st/serial v1.6.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/host/v3 v3.8.0
)
