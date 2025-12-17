package peer

import (
	"sync"

	"github.com/bashar-515/project-banana/src/backend/lib/data/errors"
	"github.com/bashar-515/project-banana/src/backend/pkg/conn"
)

type Peer struct {
	PeerId string

	SessionDescriptionSdp  string
	SessionDescriptionType int32
	SessionDescriptionMu   sync.Mutex

	IceCandidates   []*IceCandidate
	IceCandidatesMu sync.Mutex

	conn   *conn.Conn
	connMu sync.Mutex
}

type IceCandidate struct {
	candidate        string
	sdpMLineIndexStr string
	sdpMid           string
	usernameFragment string
}

const (
	MAX_NUM_ICE_CANDIDATES = 50
)

func NewPeer(peerId string) *Peer {
	return &Peer{
		PeerId: peerId,
	}
}


func (p *Peer) StoreIceCandidate(candidate, sdpMLineIndexStr, sdpMid, usernameFragment string) {
	p.IceCandidatesMu.Lock()
	defer p.IceCandidatesMu.Unlock()

	if len(p.IceCandidates) >= MAX_NUM_ICE_CANDIDATES {
		p.IceCandidates = p.IceCandidates[1:]
	}

	p.IceCandidates = append(p.IceCandidates, &IceCandidate{
		candidate: candidate,
		sdpMLineIndexStr: sdpMLineIndexStr,
		sdpMid: sdpMid,
		usernameFragment: usernameFragment,
	})
}

func (p *Peer) SendIceCandidate(candidate, sdpMLineIndexStr, sdpMid, usernameFragment string) error {
	p.connMu.Lock()
	defer p.connMu.Unlock()

	if p.conn == nil  {
		return errors.ErrNoConnection
	}

	message := map[string]any{
		"message": "candidate",
		"candidate": map[string]any{
			"candidate":        candidate,
			"sdpMLineIndex": sdpMLineIndexStr,
			"sdpMid":           sdpMid,
			"usernameFragment": usernameFragment,
		},
	}

	return p.conn.WriteJSON(message)
}

func (p *Peer) SendIceCandidates(iceCandidates []*IceCandidate) error {
	p.connMu.Lock()
	defer p.connMu.Unlock()

	if p.conn == nil {
		return errors.ErrNoConnection
	}

	candidates := make([]map[string]any, 0, len(iceCandidates))

	for _, iceCandidate := range iceCandidates {
		candidates = append(candidates, map[string]any{
			"candidate": iceCandidate.candidate,
			"sdpMLineIndex": iceCandidate.sdpMLineIndexStr,
			"sdpMid": iceCandidate.sdpMid,
			"usernameFragment": iceCandidate.usernameFragment,
		})
	}

	message := map[string]any{
		"message": "candidates",
		"candidates": candidates,
	}

	return p.conn.WriteJSON(message)
}

func (p *Peer) SetSessionDescription(sessionDescriptionSdp string, sessionDescriptionType int32) {
	p.SessionDescriptionMu.Lock()
	defer p.SessionDescriptionMu.Unlock()

	p.SessionDescriptionSdp = sessionDescriptionSdp
	p.SessionDescriptionType = sessionDescriptionType
}

func (p *Peer) SendSessionDescription(sessionDescriptionSdp string, sessionDescriptionType int32) error {
	p.connMu.Lock()
	defer p.connMu.Unlock()

	if p.conn == nil {
		return errors.ErrNoConnection
	}

	message := map[string]any{
		"message": "description",
		"description": map[string]any{
			"sdp":  sessionDescriptionSdp,
			"type": sessionDescriptionType,
		},
	}

	return p.conn.WriteJSON(message)
}

func (p *Peer) SetConn(conn *conn.Conn) {
	p.connMu.Lock()
	defer p.connMu.Unlock()

	p.conn = conn
}
