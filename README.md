# gRPoolC

A simple library to allow for pooling of gRPC connections in a way that takes advantage of service endpoints that horizontally scale.  Connections are created by a function passed to the pool, so you can define whatever options or logic you'd like.  To ensure that connections are spread across hosts behind a load balancer as they expand or contract, 5% of Get() requests will close an existing connection, and open up a new one.  Maximum connection limits are enforced by not putting excess connections back in the pool.

## Usage 

```go
import (
	"github.com/DMXRoid/grpoolc"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"fmt"
	"path.to/myrpc"
)

func setupPool() {
	grpcAddr := fmt.Sprintf("%s:%d","foo.default.svc", 8080)
	dialer := func(addr string, timeout time.Duration) (net.Conn, error) {
		return tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: true,
		})
	}
	newConnectionGenerator := func() (*grpc.ClientConn, error) {
		return grpc.Dial(*contentGrpcAddr, grpc.WithDialer(dialer), grpc.WithInsecure())
	}
	grpoolc.New("my-service", newConnectionGenerator, 100)

}

func callGrpcService() {
	conn, err := grpoolc.Get("my-service")
	defer grpoolc.Put("my-service", conn)
	ctx := context.Background()
	client := myrpc.NewMyRpcClient(conn)
	resp, err := client.MyRpc(ctx, &myrpc.MyRpcMessage{})
}

```
