package datastore

import (
	"sync"

	"github.com/bashar-515/project-banana/src/backend/lib/data/errors"
	"github.com/bashar-515/project-banana/src/backend/lib/data/models/room"
	"github.com/bashar-515/project-banana/src/backend/pkg/conn"
	"github.com/bashar-515/project-banana/src/backend/pkg/idgen"
)

type Datastore struct {
	rooms   map[string]*room.Room
	roomsMu sync.Mutex

	idGen *idgen.IdGenerator
}

func NewDatastore() *Datastore {
	return &Datastore{
		rooms: make(map[string]*room.Room),
		idGen: idgen.NewIdGenerator(),
	}
}

func (ds *Datastore) CreateRoom() (string, error) {
	roomId := ds.idGen.GenerateId()
	room, err := room.NewRoom(roomId)
	if err != nil {
		return "", err
	}

	ds.roomsMu.Lock()
	ds.rooms[roomId] = room
	ds.roomsMu.Unlock()

	return roomId, nil
}

func (ds *Datastore) AddNewPeerToRoom(roomId string) (string, error) {
	ds.roomsMu.Lock()
	room, ok := ds.rooms[roomId]
	ds.roomsMu.Unlock()
	if !ok {
		return "", errors.ErrRoomNotFound
	}

	peerId := room.AddNewPeer()

	return peerId, nil
}


func (ds *Datastore) StoreOrSendIceCandidateForPeerInRoom(roomId, peerId, candidate, sdpMLineIndexStr, sdpMid, usernameFragment string) error {
	ds.roomsMu.Lock()
	room, ok := ds.rooms[roomId]
	ds.roomsMu.Unlock()
	if !ok {
		return errors.ErrRoomNotFound
	}

	return room.StoreOrSendIceCandidateForPeer(peerId, candidate, sdpMLineIndexStr, sdpMid, usernameFragment)
}

func (ds *Datastore) SetOrSendSessionDescriptionForPeerInRoom(roomId, peerId, sessionDescriptionSdp string, sessionDescriptionType int32) error {
	ds.roomsMu.Lock()
	room, ok := ds.rooms[roomId]
	ds.roomsMu.Unlock()
	if !ok {
		return errors.ErrRoomNotFound
	}

	return room.SetOrSendSessionDescriptionForPeer(peerId, sessionDescriptionSdp, sessionDescriptionType)
}

func (ds *Datastore) SetConnForPeerInRoom(roomId, peerId string, conn *conn.Conn) error {
	ds.roomsMu.Lock()
	room, ok := ds.rooms[roomId]
	ds.roomsMu.Unlock()
	if !ok {
		return errors.ErrRoomNotFound
	}

	return room.SetConnForPeer(peerId, conn)
}

func (ds *Datastore) FlushSessionDescriptionForPeerInRoom(roomId, peerId string) error {
	ds.roomsMu.Lock()
	room, ok := ds.rooms[roomId]
	ds.roomsMu.Unlock()
	if !ok {
		return errors.ErrRoomNotFound
	}

	return room.FlushSessionDescriptionForPeer(peerId)
}

func (ds *Datastore) FlushIceCandidatesForPeerinRoom(roomId, peerId string) error {
	ds.roomsMu.Lock()
	room, ok := ds.rooms[roomId]
	ds.roomsMu.Unlock()
	if !ok {
		return errors.ErrRoomNotFound
	}

	return room.FlushIceCandidatesForPeer(peerId)
}

func (ds *Datastore) RemovePeerFromRoom(roomId, peerId string) error {
	ds.roomsMu.Lock()
	defer ds.roomsMu.Unlock()

	room, ok := ds.rooms[roomId]
	if !ok {
		return errors.ErrRoomNotFound
	}

	shouldDeleteRoom := room.RemovePeer(peerId)

	if shouldDeleteRoom {
		delete(ds.rooms, roomId)
	}

	return nil
}
