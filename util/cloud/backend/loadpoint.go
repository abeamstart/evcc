package backend

import (
	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/core/loadpoint"
	"github.com/evcc-io/evcc/core/site"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/cloud"
)

type Site struct {
	log    *util.Logger
	sender Executor
}

var _ site.API = (*Site)(nil)

func NewSite(log *util.Logger, sender Executor, id int) *Site {
	return &Site{
		log:    log,
		sender: sender,
	}
}

func (site *Site) send(api cloud.ApiCall, req *pb.BackendRequest) (*pb.EdgeResponse, error) {
	if req == nil {
		req = new(pb.BackendRequest)
	}

	req.Api = int32(api)

	resp, err := site.sender.Execute(req)
	if err != nil {
		site.log.ERROR.Printf("calling %d: %v", api, err)
	}

	return resp, err
}

func (site *Site) Healthy() bool {
	resp, err := site.send(cloud.Healthy, nil)
	if err != nil {
		return false
	}
	return resp.Payload.BoolVal
}

func (site *Site) LoadPoints() []loadpoint.API {
	// TODO
	return nil
}

func (site *Site) SetPrioritySoC(soc float64) error {
	_, err := site.send(cloud.SetPrioritySoC, &pb.BackendRequest{
		Payload: &pb.Payload{
			FloatVal: float64(soc),
		},
	})

	return err
}
