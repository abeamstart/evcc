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
	go sendUpdates(client, in)

	stream, err := client.SubscribeBackendRequest(context.Background(), new(pb.SubscribeRequest))
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go receiveRequest(stream, site, done)

	return nil
}

func sendUpdates(client pb.CloudConnectServiceClient, in <-chan util.Param) {
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

		ctx := context.Background()
		if _, err := client.SendEdgeUpdate(ctx, &req); err != nil {
			panic(err)
		}
	}
}

func receiveRequest(stream pb.CloudConnectService_SubscribeBackendRequestClient, site site.API, done chan struct{}) {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			close(done)
			return
		}

		if err != nil {
			fmt.Println("cannot receive", err)
			os.Exit(1)
		}

		handleRequest(site, req)
	}
}

func handleRequest(site site.API, req *pb.BackendRequest) {
}
