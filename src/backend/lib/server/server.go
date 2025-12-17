package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/gorilla/websocket"

	"github.com/bashar-515/project-banana/gen/go/app/v1/appv1connect"
	appv1 "github.com/bashar-515/project-banana/gen/go/app/v1"

	"github.com/bashar-515/project-banana/src/backend/lib/data/datastore"
	"github.com/bashar-515/project-banana/src/backend/pkg/conn"
)

var errNoMsg = errors.New("no message")

type Server struct {
	appv1connect.UnimplementedAppServiceHandler

	store *datastore.Datastore
}

func NewServer() *Server {
	return &Server{
		store: datastore.NewDatastore(),
	}
}

func (s *Server) CreateRoom(
	ctx context.Context,
	req *connect.Request[appv1.CreateRoomRequest],
) (*connect.Response[appv1.CreateRoomResponse], error) {
	roomId, err := s.store.CreateRoom()
	if err != nil {
		return nil, err
	}

	createRoomResponse := &appv1.CreateRoomResponse{
		RoomId: roomId,
	}

	response := connect.NewResponse(createRoomResponse)

	return response, nil
}

func (s *Server) JoinRoom(
	ctx context.Context,
	req *connect.Request[appv1.JoinRoomRequest],
) (*connect.Response[appv1.JoinRoomResponse], error) {
	if req.Msg == nil {
		return nil, errNoMsg
	}

	roomId := req.Msg.GetRoomId()
	peerId, err := s.store.AddNewPeerToRoom(roomId)
	if err != nil {
		return nil, err
	}

	joinRoomResponse := &appv1.JoinRoomResponse{
		PeerId: peerId,
	}

	response := connect.NewResponse(joinRoomResponse)

	return response, nil
}

func (s *Server) UploadIceCandidate(
	ctx context.Context,
	req *connect.Request[appv1.UploadIceCandidateRequest],
) (*connect.Response[appv1.UploadIceCandidateResponse], error) {
	if req.Msg == nil {
		return nil, errNoMsg
	}

	roomId := req.Msg.GetRoomId()
	peerId := req.Msg.GetPeerId()
	iceCandidate := req.Msg.GetIceCandidate()

	var candidate string
	if iceCandidate.Candidate != nil {
		candidate = *iceCandidate.Candidate
	}

	var sdpMLineIndexStr string
	if iceCandidate.SdpMLineIndex != nil {
		sdpMLineIndexStr = strconv.Itoa(int(*iceCandidate.SdpMLineIndex))
	}

	var sdpMid string
	if iceCandidate.SdpMid != nil {
		sdpMid = *iceCandidate.SdpMid
	}

	var usernameFragment string
	if iceCandidate.UsernameFragment != nil {
		usernameFragment = *iceCandidate.UsernameFragment
	}

	err := s.store.StoreOrSendIceCandidateForPeerInRoom(
		roomId,
		peerId,
		candidate,
		sdpMLineIndexStr,
		sdpMid,
		usernameFragment,
	)
	if err != nil {
		return nil, err
	}

	uploadIceCandidateResponse := &appv1.UploadIceCandidateResponse{}
	response := connect.NewResponse(uploadIceCandidateResponse)

	return response, nil
}

func (s *Server) UploadSessionDescription(
	ctx context.Context,
	req *connect.Request[appv1.UploadSessionDescriptionRequest],
) (*connect.Response[appv1.UploadSessionDescriptionResponse], error) {
	if req.Msg == nil {
		return nil, errNoMsg
	}

	roomId := req.Msg.GetRoomId()
	peerId := req.Msg.GetPeerId()
	sessionDescription := req.Msg.GetSessionDescription()

	var sdp string
	if sessionDescription.Sdp != nil {
		sdp = *sessionDescription.Sdp
	}

	err := s.store.SetOrSendSessionDescriptionForPeerInRoom(roomId, peerId, sdp, int32(sessionDescription.Type))
	if err != nil {
		return nil, err
	}

	uploadSessionDescriptionResponse := &appv1.UploadSessionDescriptionResponse{}
	response := connect.NewResponse(uploadSessionDescriptionResponse)

	return response, nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://app.project-banana.com"
	},
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	roomId := query.Get("roomId")
	peerId := query.Get("peerId")

	if roomId == "" && peerId == "" {
		http.Error(w, "missing room and peer ID's", http.StatusBadRequest)
		return
	} else if roomId == "" {
		http.Error(w, "missing room ID", http.StatusBadRequest)
		return
	} else if peerId == "" {
		http.Error(w, "missing peer ID", http.StatusBadRequest)
		return
	}

	websocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "error upgrading connection", http.StatusInternalServerError)
		return
	}

	conn := conn.NewConn(websocket)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf(err.Error())
		}
	}()

	err = s.store.SetConnForPeerInRoom(roomId, peerId, conn)
	if err != nil {
		return
	}

	err = s.store.FlushSessionDescriptionForPeerInRoom(roomId, peerId)
	if err != nil {
		return
	}

	err = s.store.FlushIceCandidatesForPeerinRoom(roomId, peerId)
	if err != nil {
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err == nil {
			continue
		}

		err = s.store.RemovePeerFromRoom(roomId, peerId)
		if err != nil {
			log.Print(err.Error())
		}

		return
	}
}
