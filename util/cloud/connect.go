package cloud

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/util"
	"google.golang.org/grpc"
)

func Connect(conn *grpc.ClientConn, in <-chan util.Param) {
	// create stream
	client := pb.NewCloudConnectServiceClient(conn)

	go func() {
		var b bytes.Buffer
		enc := gob.NewEncoder(&b)

		for param := range in {
			b.Reset()
			if err := enc.Encode(param.Val); err != nil {
				panic(err)
			}

			fmt.Println("--")
			fmt.Printf("%d %0x\n", b.Len(), b.Bytes())

			var lp int32
			if param.LoadPoint != nil {
				lp = int32(*param.LoadPoint + 1)
			}

			req := pb.UpdateRequest{
				Loadpoint: lp,
				Key:       param.Key,
				Val:       b.Bytes(),
			}

			// check ----
			res := util.Param{
				Key: req.Key,
			}

			if req.Loadpoint > 0 {
				i := int(req.Loadpoint - 1)
				res.LoadPoint = &i
			}

			err := gob.NewDecoder(bytes.NewBuffer(req.Val)).Decode(res.Val)
			if err != nil {
				panic(err)
			}
			// check ----

			ctx := context.Background()
			if _, err := client.SendEdgeUpdate(ctx, &req); err != nil {
				panic(err)
			}
		}
	}()

	stream, err := client.SubscribeBackendRequest(context.Background(), new(pb.SubscribeRequest))
	if err != nil {
		fmt.Println("open stream error", err)
		os.Exit(1)
	}

	// done := make(chan struct{})

	go func() {
		for {
			resp, err := stream.Recv()

			// if err == io.EOF {
			// 	close(done)
			// 	return
			// }

			if err != nil {
				fmt.Println("cannot receive", err)
				os.Exit(1)
			}

			fmt.Println("Resp received:", resp.Result)
		}
	}()

	// <-done
}
