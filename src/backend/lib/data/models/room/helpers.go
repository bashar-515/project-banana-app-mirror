package room

import "github.com/bashar-515/project-banana/src/backend/lib/data/models/peer"

func (r *Room) getOtherPeer(peerId string) *peer.Peer {
	r.peersMu.Lock()
	defer r.peersMu.Unlock()

	if len(r.peers) <= 1 {
		return nil
	}
	_, ok := r.peers[peerId]
	if !ok {
		return nil
	}

	for id, peer := range r.peers {
		if id != peerId {
			return peer
		}
	}

	return nil
}
