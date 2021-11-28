package cloud

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/evcc-io/evcc/api/proto/pb"
	"github.com/evcc-io/evcc/util"
	"google.golang.org/grpc"
)

func Connect(conn *grpc.ClientConn, in <-chan util.Param) {
	// create stream
	client := pb.NewStreamServiceClient(conn)

	stream, err := client.FetchResponse(context.Background(), &pb.Request{Id: 1})
	if err != nil {
		fmt.Println("open stream error", err)
		os.Exit(1)
	}

	done := make(chan struct{})

	go func() {
		for {
			resp, err := stream.Recv()

			if err == io.EOF {
				close(done)
				return
			}

			if err != nil {
				fmt.Println("cannot receive", err)
				os.Exit(1)
			}

			fmt.Println("Resp received:", resp.Result)
		}
	}()

	<-done
	fmt.Println("finished")
}
