package cloud

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/core"
	"github.com/evcc-io/evcc/core/loadpoint"
	"github.com/evcc-io/evcc/core/site"
	"github.com/evcc-io/evcc/util"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

func init() {
	RegisterTypes()
}

func RegisterTypes() {
	gob.Register(api.ModeEmpty)
	gob.Register(api.StatusNone)
	gob.Register(loadpoint.RemoteEnable)
	gob.Register(time.Duration(0))
	gob.Register(time.Time{})
}

func Connect(conn *grpc.ClientConn, site *core.Site, in <-chan util.Param) error {
	client := pb.NewCloudConnectServiceClient(conn)

	updateS, err := client.SendEdgeUpdate(context.Background())
	if err != nil {
		return err
	}

	go sendUpdates(updateS, in)

	inS, err := client.SubscribeEdgeRequest(context.Background(), new(pb.Empty))
	if err != nil {
		return err
	}

	outS, err := client.SendEdgeResponse(context.Background())
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go handleRequest(inS, outS, site, done)

	return nil
}

func sendUpdates(outS pb.CloudConnectService_SendEdgeUpdateClient, in <-chan util.Param) {
	b := new(bytes.Buffer)

	for param := range in {
		enc := gob.NewEncoder(b)

		b.Reset()
		if err := enc.Encode(&param.Val); err != nil {
			panic(err)
		}

		var lp int32
		if param.LoadPoint != nil {
			lp = int32(*param.LoadPoint + 1)
		}

		req := pb.UpdateRequest{
			Loadpoint: lp,
			Key:       param.Key,
			Val:       b.Bytes(),
		}

		if err := outS.Send(&req); err != nil {
			panic(err)
		}
	}
}

func handleRequest(inS pb.CloudConnectService_SubscribeEdgeRequestClient, outS pb.CloudConnectService_SendEdgeResponseClient, site site.API, done chan struct{}) {
	for {
		req, err := inS.Recv()
		if err == io.EOF {
			close(done)
			return
		}

		if err != nil {
			fmt.Println("cannot receive", err)
			os.Exit(1)
		}

		resp, err := processRequest(site, req)
		if err != nil {
			resp.Error = err.Error()
		}

		if err := outS.Send(resp); err != nil {
			panic(err)
		}
	}
}

func processRequest(site site.API, req *pb.EdgeRequest) (*pb.EdgeResponse, error) {
	res := &pb.EdgeResponse{
		Id: req.Id,
	}

	var lp loadpoint.API
	if req.Loadpoint > 0 {
		lp = site.LoadPoints()[req.Loadpoint-1]
	}

	var err error

	switch ApiCall(req.Api) {
	case Name:
		res.Stringval = lp.Name()

	case HasChargeMeter:
		res.Boolval = lp.HasChargeMeter()

	case GetStatus:
		res.Stringval = string(lp.GetStatus())

	case GetMode:
		res.Stringval = string(lp.GetMode())

	case SetMode:
		lp.SetMode(api.ChargeMode(req.Stringval))

	case GetTargetSoC:
		res.Intval = int64(lp.GetTargetSoC())

	case SetTargetSoC:
		err = lp.SetTargetSoC(int(req.Intval))

	case GetMinSoC:
		res.Intval = int64(lp.GetMinSoC())

	case SetMinSoC:
		err = lp.SetMinSoC(int(req.Intval))

	case GetPhases:
		res.Intval = int64(lp.GetPhases())

	case SetPhases:
		err = lp.SetPhases(int(req.Intval))

	case SetTargetCharge:
		lp.SetTargetCharge(req.Timeval.AsTime(), int(req.Intval))

	case GetChargePower:
		res.Floatval = lp.GetChargePower()

	case GetMinCurrent:
		res.Floatval = lp.GetMinCurrent()

	case SetMinCurrent:
		lp.SetMinCurrent(req.Floatval)

	case GetMaxCurrent:
		res.Floatval = lp.GetMaxCurrent()

	case SetMaxCurrent:
		lp.SetMaxCurrent(req.Floatval)

	case GetMinPower:
		res.Floatval = lp.GetMinPower()

	case GetMaxPower:
		res.Floatval = lp.GetMaxPower()

	case GetRemainingDuration:
		res.Durationval = durationpb.New(lp.GetRemainingDuration())

	case GetRemainingEnergy:
		res.Floatval = lp.GetRemainingEnergy()

	case RemoteControl:
		lp.RemoteControl("my.evcc.io", loadpoint.RemoteDemand(req.Stringval))

	default:
		err = fmt.Errorf("unknown api call %d", req.Api)
	}

	return res, err
}
