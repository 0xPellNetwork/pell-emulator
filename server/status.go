package server

import "sync"

type ServerStatus struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
}

func (ss *ServerStatus) Disable(msg string) {
	ss.Ready = false
	ss.Message = msg
}

func (ss *ServerStatus) Enable() {
	ss.Ready = true
	ss.Message = "ok"
}

var (
	emulatorServerState = ServerStatus{
		Ready:   false,
		Message: "initing...",
	}
	statusMutex sync.RWMutex
)
