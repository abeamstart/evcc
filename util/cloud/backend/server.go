package backend

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"

	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/cloud"
	"google.golang.org/grpc/peer"
)

type Server struct {
	mu            sync.Mutex
	log           *util.Logger
	UpdateHandler func(util.Param)
	pb.UnimplementedCloudConnectServiceServer
	clients []*EdgeClient
}

type Sender interface {
	Send(*pb.BackendRequest) error
}

type Receiver interface {
	Receive(*pb.BackendRequest) error
}

type RoundTripper struct {
	send Sender
	recv Receiver
}

// SubscribeBackendRequest connects an edge client to the backend. The edge client will receive backend requests.
func (s *Server) SubscribeBackendRequest(req *pb.EdgeEnvironment, srv pb.CloudConnectService_SubscribeBackendRequestServer) error {
	fmt.Println("SubscribeBackendRequest")

	rt := &RoundTripper{
		send: srv,
		recv: nil,
	}

	peer, ok := peer.FromContext(srv.Context())
	if !ok {
		return errors.New("missing peer info")
	}
	fmt.Println("peer:", peer)

	client := NewEdgeClient(req, rt, peer)

	srv.Send(&pb.BackendRequest{
		Id:        1,
		Api:       int32(cloud.Name),
		Loadpoint: 1,
	})

	s.mu.Lock()
	s.clients = append(s.clients, client)
	s.mu.Unlock()

	return nil
}

// SendEdgeResponse receives edge client responses in reply to backend requests.
func (s *Server) SendEdgeResponse(inS pb.CloudConnectService_SendEdgeResponseServer) error {
	for {
		fmt.Println("SendEdgeResponse")

		req, err := inS.Recv()
		if err != nil {
			fmt.Println("SendEdgeResponse", err)
			return err
		}

		fmt.Println("SendEdgeResponse", req)

		p, ok := peer.FromContext(inS.Context())
		if !ok {
			return errors.New("missing peer info")
		}

		_ = p
		_ = req
	}
}

func ParamFromUpdateRequest(req *pb.UpdateRequest) (util.Param, error) {
	param := util.Param{
		Key: req.Key,
	}

	if req.Loadpoint > 0 {
		i := int(req.Loadpoint - 1)
		param.LoadPoint = &i
	}

	err := gob.NewDecoder(bytes.NewReader(req.Val)).Decode(&param.Val)

	return param, err
}

func (s *Server) SendEdgeUpdate(inS pb.CloudConnectService_SendEdgeUpdateServer) error {
	for {
		req, err := inS.Recv()
		if err != nil {
			return err
		}

		param, err := ParamFromUpdateRequest(req)
		if err != nil {
			s.log.ERROR.Printf("failed to decode update request: %v", err)
		} else {
			s.UpdateHandler(param)
		}
	}
}
