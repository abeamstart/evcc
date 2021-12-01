package edge

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"os"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/core"
	"github.com/evcc-io/evcc/core/loadpoint"
	"github.com/evcc-io/evcc/core/site"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/cloud"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ConnectToBackend(conn *grpc.ClientConn, site *core.Site, in <-chan util.Param) error {
	client := pb.NewCloudConnectServiceClient(conn)

	// edge to backend

	updateS, err := client.SendEdgeUpdate(context.Background())
	if err != nil {
		return err
	}

	go sendUpdates(updateS, in)

	// backend to edge

	req := &pb.EdgeEnvironment{
		Loadpoints: int32(len(site.LoadPoints())),
	}

	inS, err := client.SubscribeEdgeRequest(context.Background(), req)
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

		resp, err := apiRequest(site, req)
		if err != nil {
			resp.Error = err.Error()
		}

		if err := outS.Send(resp); err != nil {
			panic(err)
		}
	}
}

func apiRequest(site site.API, req *pb.EdgeRequest) (*pb.EdgeResponse, error) {
	res := &pb.EdgeResponse{
		Id: req.Id,
	}

	var lp loadpoint.API
	if req.Loadpoint > 0 {
		lp = site.LoadPoints()[req.Loadpoint-1]
	}

	var err error

	switch cloud.ApiCall(req.Api) {
	case cloud.Name:
		res.Payload.StringVal = lp.Name()

	case cloud.HasChargeMeter:
		res.Payload.BoolVal = lp.HasChargeMeter()

	case cloud.GetStatus:
		res.Payload.StringVal = string(lp.GetStatus())

	case cloud.GetMode:
		res.Payload.StringVal = string(lp.GetMode())

	case cloud.SetMode:
		lp.SetMode(api.ChargeMode(req.Payload.StringVal))

	case cloud.GetTargetSoC:
		res.Payload.IntVal = int64(lp.GetTargetSoC())

	case cloud.SetTargetSoC:
		lp.SetTargetSoC(int(req.Payload.IntVal))

	case cloud.GetMinSoC:
		res.Payload.IntVal = int64(lp.GetMinSoC())

	case cloud.SetMinSoC:
		lp.SetMinSoC(int(req.Payload.IntVal))

	case cloud.GetPhases:
		res.Payload.IntVal = int64(lp.GetPhases())

	case cloud.SetPhases:
		err = lp.SetPhases(int(req.Payload.IntVal))

	case cloud.SetTargetCharge:
		lp.SetTargetCharge(req.Payload.TimeVal.AsTime(), int(req.Payload.IntVal))

	case cloud.GetChargePower:
		res.Payload.FloatVal = lp.GetChargePower()

	case cloud.GetMinCurrent:
		res.Payload.FloatVal = lp.GetMinCurrent()

	case cloud.SetMinCurrent:
		lp.SetMinCurrent(req.Payload.FloatVal)

	case cloud.GetMaxCurrent:
		res.Payload.FloatVal = lp.GetMaxCurrent()

	case cloud.SetMaxCurrent:
		lp.SetMaxCurrent(req.Payload.FloatVal)

	case cloud.GetMinPower:
		res.Payload.FloatVal = lp.GetMinPower()

	case cloud.GetMaxPower:
		res.Payload.FloatVal = lp.GetMaxPower()

	case cloud.GetRemainingDuration:
		res.Payload.DurationVal = durationpb.New(lp.GetRemainingDuration())

	case cloud.GetRemainingEnergy:
		res.Payload.FloatVal = lp.GetRemainingEnergy()

	case cloud.RemoteControl:
		lp.RemoteControl("my.evcc.io", loadpoint.RemoteDemand(req.Payload.StringVal))

	default:
		err = fmt.Errorf("unknown api call %d", req.Api)
	}

	return res, err
}
