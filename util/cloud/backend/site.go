package backend

import (
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/core/loadpoint"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/cloud"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Loadpoint struct {
	log    *util.Logger
	sender Executor
	ID     int
}

type Executor interface {
	Execute(*pb.BackendRequest) (*pb.EdgeResponse, error)
}

var _ loadpoint.API = (*Loadpoint)(nil)

func NewLoadpoint(log *util.Logger, sender Executor, id int) *Loadpoint {
	return &Loadpoint{
		log:    log,
		sender: sender,
		ID:     id,
	}
}

func (lp *Loadpoint) send(api cloud.ApiCall, req *pb.BackendRequest) (*pb.EdgeResponse, error) {
	if req == nil {
		req = new(pb.BackendRequest)
	}

	req.Loadpoint = int32(lp.ID + 1)
	req.Api = int32(api)

	resp, err := lp.sender.Execute(req)
	if err != nil {
		lp.log.ERROR.Printf("calling %d: %v", api, err)
	}

	return resp, err
}

func (lp *Loadpoint) Name() string {
	resp, err := lp.send(cloud.Name, nil)
	if err != nil {
		return ""
	}
	return resp.Payload.StringVal
}

func (lp *Loadpoint) HasChargeMeter() bool {
	resp, err := lp.send(cloud.HasChargeMeter, nil)
	if err != nil {
		return false
	}
	return resp.Payload.BoolVal

}

func (lp *Loadpoint) GetStatus() api.ChargeStatus {
	resp, err := lp.send(cloud.GetStatus, nil)
	if err != nil {
		return api.StatusNone
	}
	return api.ChargeStatus(resp.Payload.StringVal)
}

func (lp *Loadpoint) GetMode() api.ChargeMode {
	resp, err := lp.send(cloud.GetMode, nil)
	if err != nil {
		return api.ModeEmpty
	}
	return api.ChargeMode(resp.Payload.StringVal)
}

func (lp *Loadpoint) SetMode(val api.ChargeMode) {
	_, _ = lp.send(cloud.RemoteControl, &pb.BackendRequest{Payload: &pb.Payload{StringVal: string(val)}})
}

func (lp *Loadpoint) GetTargetSoC() int {
	resp, err := lp.send(cloud.GetTargetSoC, nil)
	if err != nil {
		return 0
	}
	return int(resp.Payload.IntVal)
}

func (lp *Loadpoint) SetTargetSoC(val int) {
	_, _ = lp.send(cloud.SetTargetSoC, &pb.BackendRequest{Payload: &pb.Payload{IntVal: int64(val)}})
}

func (lp *Loadpoint) GetMinSoC() int {
	resp, err := lp.send(cloud.GetMinSoC, nil)
	if err != nil {
		return 0
	}
	return int(resp.Payload.IntVal)
}

func (lp *Loadpoint) SetMinSoC(val int) {
	_, _ = lp.send(cloud.SetMinSoC, &pb.BackendRequest{Payload: &pb.Payload{IntVal: int64(val)}})
}

func (lp *Loadpoint) GetPhases() int {
	resp, err := lp.send(cloud.GetPhases, nil)
	if err != nil {
		return 0
	}
	return int(resp.Payload.IntVal)
}

func (lp *Loadpoint) SetPhases(val int) error {
	_, err := lp.send(cloud.SetPhases, &pb.BackendRequest{Payload: &pb.Payload{IntVal: int64(val)}})
	return err
}

func (lp *Loadpoint) SetTargetCharge(t time.Time, val int) {
	_, err := lp.send(cloud.SetTargetCharge, &pb.BackendRequest{Payload: &pb.Payload{
		TimeVal: timestamppb.New(t),
		IntVal:  int64(val),
	}})
	_ = err
}

func (lp *Loadpoint) GetChargePower() float64 {
	resp, err := lp.send(cloud.GetChargePower, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.FloatVal
}

func (lp *Loadpoint) GetMinCurrent() float64 {
	resp, err := lp.send(cloud.GetMinCurrent, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.FloatVal
}

func (lp *Loadpoint) SetMinCurrent(val float64) {
	_, err := lp.send(cloud.SetMinCurrent, &pb.BackendRequest{Payload: &pb.Payload{FloatVal: val}})
	_ = err
}

func (lp *Loadpoint) GetMaxCurrent() float64 {
	resp, err := lp.send(cloud.GetMaxCurrent, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.FloatVal
}

func (lp *Loadpoint) SetMaxCurrent(val float64) {
	_, err := lp.send(cloud.SetMaxCurrent, &pb.BackendRequest{Payload: &pb.Payload{FloatVal: val}})
	_ = err
}

func (lp *Loadpoint) GetMinPower() float64 {
	resp, err := lp.send(cloud.GetMinPower, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.FloatVal
}

func (lp *Loadpoint) GetMaxPower() float64 {
	resp, err := lp.send(cloud.GetMaxPower, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.FloatVal
}

func (lp *Loadpoint) GetRemainingDuration() time.Duration {
	resp, err := lp.send(cloud.GetRemainingDuration, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.DurationVal.AsDuration()
}

func (lp *Loadpoint) GetRemainingEnergy() float64 {
	resp, err := lp.send(cloud.GetRemainingEnergy, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.FloatVal
}

func (lp *Loadpoint) RemoteControl(_ string, demand loadpoint.RemoteDemand) {
	_, _ = lp.send(cloud.RemoteControl, &pb.BackendRequest{Payload: &pb.Payload{StringVal: string(demand)}})
}
