package errors

import "errors"

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrPeerNotFound = errors.New("peer not found")
	ErrNoConnection = errors.New("no connection")
)
