package mocks

import "github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/chainflags"

var KeyFileFlag = chainflags.StringFlag{
	Name:  "key-file",
	Usage: "key file",
}
