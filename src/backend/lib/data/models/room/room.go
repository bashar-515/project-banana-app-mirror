package room

import (
	"errors"
	"sync"

	errorslib "github.com/bashar-515/project-banana/src/backend/lib/data/errors"
	"github.com/bashar-515/project-banana/src/backend/lib/data/models/peer"
	"github.com/bashar-515/project-banana/src/backend/pkg/conn"
	"github.com/bashar-515/project-banana/src/backend/pkg/idgen"
)

type Room struct {
	roomId string

	peers   map[string]*peer.Peer
	peersMu sync.Mutex

	idGen *idgen.IdGenerator
}

func NewRoom(id string) (*Room, error) {
	idGen, err := idgen.NewIdGeneratorWithNamespace(id)
	if err != nil {
		return nil, err
	}

	room := &Room{
		roomId: id,
		peers: make(map[string]*peer.Peer, 2),
		idGen: idGen,
	}

	return room, nil
}

func (r *Room) AddNewPeer() string {
	peerId := r.idGen.GenerateId()
	peer := peer.NewPeer(peerId)

	r.peersMu.Lock()
	r.peers[peerId] = peer
	r.peersMu.Unlock()

	return peerId
}

func (r *Room) StoreOrSendIceCandidateForPeer(peerId, candidate, sdpMLineIndexStr, sdpMid, usernameFragment string) error {
	r.peersMu.Lock()
	peer, ok := r.peers[peerId]
	r.peersMu.Unlock()
	if !ok {
		return errorslib.ErrPeerNotFound
	}

	otherPeer := r.getOtherPeer(peerId)

	if otherPeer == nil {
		peer.StoreIceCandidate(candidate, sdpMLineIndexStr, sdpMid, usernameFragment)

		return nil
	}

	err := otherPeer.SendIceCandidate(candidate, sdpMLineIndexStr, sdpMid, usernameFragment)
	if err != nil {
		if errors.Is(err, errorslib.ErrNoConnection) {
			peer.StoreIceCandidate(candidate, sdpMLineIndexStr, sdpMid, usernameFragment)
		}

		return err
	}

	return nil
}

func (r *Room) SetOrSendSessionDescriptionForPeer(peerId, sessionDescriptionSdp string, sessionDescriptionType int32) error {
	r.peersMu.Lock()
	peer, ok := r.peers[peerId]
	r.peersMu.Unlock()
	if !ok {
		return errorslib.ErrPeerNotFound
	}

	otherPeer := r.getOtherPeer(peerId)

	if otherPeer == nil {
		peer.SetSessionDescription(sessionDescriptionSdp, sessionDescriptionType)

		return nil
	}

	err := otherPeer.SendSessionDescription(sessionDescriptionSdp, sessionDescriptionType)
	if err != nil {
		if errors.Is(err, errorslib.ErrNoConnection) {
			peer.SetSessionDescription(sessionDescriptionSdp, sessionDescriptionType)	
		}

		return err
	}

	return nil
}

func (r *Room) SetConnForPeer(peerId string, conn *conn.Conn) error {
	r.peersMu.Lock()
	peer, ok := r.peers[peerId]
	r.peersMu.Unlock()
	if !ok {
		return errorslib.ErrPeerNotFound
	}

	peer.SetConn(conn)

	return nil
}

func (r *Room) FlushSessionDescriptionForPeer(peerId string) error {
	r.peersMu.Lock()
	toPeer, ok := r.peers[peerId]
	r.peersMu.Unlock()
	if !ok {
		return errorslib.ErrPeerNotFound
	}

	fromPeer := r.getOtherPeer(peerId)

	if fromPeer == nil {
		return nil
	}

	fromPeer.SessionDescriptionMu.Lock()
	sessionDescriptionSdp := fromPeer.SessionDescriptionSdp
	sessionDescriptionType := fromPeer.SessionDescriptionType
	fromPeer.SessionDescriptionMu.Unlock()

	return toPeer.SendSessionDescription(sessionDescriptionSdp, sessionDescriptionType)
}

func (r *Room) FlushIceCandidatesForPeer(peerId string) error {
	r.peersMu.Lock()
	toPeer, ok := r.peers[peerId]
	r.peersMu.Unlock()
	if !ok {
		return errorslib.ErrPeerNotFound
	}

	fromPeer := r.getOtherPeer(peerId)

	if fromPeer == nil {
		return nil
	}

	var iceCandidates []*peer.IceCandidate

	for {
		var iceCandidate *peer.IceCandidate

		fromPeer.IceCandidatesMu.Lock()
		if len(fromPeer.IceCandidates) > 0 {
			iceCandidate = fromPeer.IceCandidates[0]
			fromPeer.IceCandidates = fromPeer.IceCandidates[:1]
		}
		fromPeer.IceCandidatesMu.Unlock()

		if iceCandidate == nil {
			break
		}

		iceCandidates = append(iceCandidates, iceCandidate)
	}

	return toPeer.SendIceCandidates(iceCandidates)
}

func (r *Room) RemovePeer(peerId string) bool {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	delete(r.peers, peerId)

	return len(r.peers) == 0
}
