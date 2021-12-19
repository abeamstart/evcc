package backend

import (
	"errors"
	"fmt"

	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/util"
	"google.golang.org/grpc/peer"
)

type EdgeClient struct {
	id         string
	loadpoints []*Adapter
	conn       pb.CloudConnectService_SubscribeBackendRequestServer
	rt         *RoundTripper
	peer       *peer.Peer
}

func NewEdgeClient(req *pb.EdgeEnvironment, rt *RoundTripper, peer *peer.Peer) *EdgeClient {
	c := &EdgeClient{
		id: req.GetEdgeId(),
		// conn: conn,
		rt:   rt,
		peer: peer,
	}

	for i := 1; i <= int(req.GetLoadpoints()); i++ {
		log := util.NewLogger(fmt.Sprintf("lp-%d", i))
		c.loadpoints = append(c.loadpoints, NewAdapter(log, c, i))
	}

	return c
}

func (c *EdgeClient) Execute(req *pb.BackendRequest) (*pb.EdgeResponse, error) {
	err := c.conn.Send(req)
	if err != nil {
		return nil, err
	}

	return nil, errors.New("unimplemented")
}
