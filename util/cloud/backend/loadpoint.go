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

type Adapter struct {
	log    *util.Logger
	sender Sender
	ID     int32
}

type Sender interface {
	Execute(*pb.EdgeRequest) (*pb.EdgeResponse, error)
}

var _ loadpoint.API = (*Adapter)(nil)

func NewAdapter(log *util.Logger, sender Sender, id int32) *Adapter {
	return &Adapter{
		log:    log,
		sender: sender,
		ID:     id,
	}
}

func (lp *Adapter) send(api cloud.ApiCall, req *pb.EdgeRequest) (*pb.EdgeResponse, error) {
	if req == nil {
		req = &pb.EdgeRequest{}
	}

	req.Loadpoint = lp.ID + 1
	resp, err := lp.sender.Execute(req)
	if err != nil {
		lp.log.ERROR.Printf("calling %d: %v", api, err)
	}

	return resp, err
}

func (lp *Adapter) Name() string {
	resp, err := lp.send(cloud.Name, nil)
	if err != nil {
		return ""
	}
	return resp.Payload.Stringval
}

func (lp *Adapter) HasChargeMeter() bool {
	resp, err := lp.send(cloud.HasChargeMeter, nil)
	if err != nil {
		return false
	}
	return resp.Payload.Boolval

}

func (lp *Adapter) GetStatus() api.ChargeStatus {
	resp, err := lp.send(cloud.GetStatus, nil)
	if err != nil {
		return ""
	}
	return api.ChargeStatus(resp.Payload.Stringval)
}

func (lp *Adapter) GetMode() api.ChargeMode {
	resp, err := lp.send(cloud.GetMode, nil)
	if err != nil {
		return ""
	}
	return api.ChargeMode(resp.Payload.Stringval)
}

func (lp *Adapter) SetMode(val api.ChargeMode) {
	_, err := lp.send(cloud.RemoteControl, &pb.EdgeRequest{Payload: &pb.Payload{Stringval: string(val)}})
	_ = err
}

func (lp *Adapter) GetTargetSoC() int {
	resp, err := lp.send(cloud.GetTargetSoC, nil)
	if err != nil {
		return 0
	}
	return int(resp.Payload.Intval)
}

func (lp *Adapter) SetTargetSoC(val int) error {
	_, err := lp.send(cloud.SetTargetSoC, &pb.EdgeRequest{Payload: &pb.Payload{Intval: int64(val)}})
	return err
}

func (lp *Adapter) GetMinSoC() int {
	resp, err := lp.send(cloud.GetMinSoC, nil)
	if err != nil {
		return 0
	}
	return int(resp.Payload.Intval)
}

func (lp *Adapter) SetMinSoC(val int) error {
	_, err := lp.send(cloud.SetMinSoC, &pb.EdgeRequest{Payload: &pb.Payload{Intval: int64(val)}})
	_ = err
	return err
}

func (lp *Adapter) GetPhases() int {
	resp, err := lp.send(cloud.GetPhases, nil)
	if err != nil {
		return 0
	}
	return int(resp.Payload.Intval)
}

func (lp *Adapter) SetPhases(val int) error {
	_, err := lp.send(cloud.SetPhases, &pb.EdgeRequest{Payload: &pb.Payload{Intval: int64(val)}})
	return err
}

func (lp *Adapter) SetTargetCharge(t time.Time, val int) {
	_, err := lp.send(cloud.SetTargetCharge, &pb.EdgeRequest{Payload: &pb.Payload{
		Timeval: timestamppb.New(t),
		Intval:  int64(val),
	}})
	_ = err
}

func (lp *Adapter) GetChargePower() float64 {
	resp, err := lp.send(cloud.GetChargePower, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.Floatval
}

func (lp *Adapter) GetMinCurrent() float64 {
	resp, err := lp.send(cloud.GetMinCurrent, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.Floatval
}

func (lp *Adapter) SetMinCurrent(val float64) {
	_, err := lp.send(cloud.SetMinCurrent, &pb.EdgeRequest{Payload: &pb.Payload{Floatval: val}})
	_ = err
}

func (lp *Adapter) GetMaxCurrent() float64 {
	resp, err := lp.send(cloud.GetMaxCurrent, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.Floatval
}

func (lp *Adapter) SetMaxCurrent(val float64) {
	_, err := lp.send(cloud.SetMaxCurrent, &pb.EdgeRequest{Payload: &pb.Payload{Floatval: val}})
	_ = err
}

func (lp *Adapter) GetMinPower() float64 {
	resp, err := lp.send(cloud.GetMinPower, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.Floatval
}

func (lp *Adapter) GetMaxPower() float64 {
	resp, err := lp.send(cloud.GetMaxPower, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.Floatval
}

func (lp *Adapter) GetRemainingDuration() time.Duration {
	resp, err := lp.send(cloud.GetRemainingDuration, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.Durationval.AsDuration()
}

func (lp *Adapter) GetRemainingEnergy() float64 {
	resp, err := lp.send(cloud.GetRemainingEnergy, nil)
	if err != nil {
		return 0
	}
	return resp.Payload.Floatval
}

func (lp *Adapter) RemoteControl(_ string, demand loadpoint.RemoteDemand) {
	_, err := lp.send(cloud.RemoteControl, &pb.EdgeRequest{Payload: &pb.Payload{Stringval: string(demand)}})
	_ = err
}
