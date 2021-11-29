package backend

import (
	"bytes"
	"encoding/gob"

	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/util"
)

type Server struct {
	log           *util.Logger
	UpdateHandler func(util.Param)
	pb.UnimplementedCloudConnectServiceServer
}

func (s *Server) SubscribeEdgeRequest(req *pb.EdgeEnvironment, srv pb.CloudConnectService_SubscribeEdgeRequestServer) error {
	// handle edge clients connecting
	return nil
}

func (s *Server) SendEdgeResponse(inS pb.CloudConnectService_SendEdgeResponseServer) error {
	for {
		req, err := inS.Recv()
		if err != nil {
			return err
		}
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
