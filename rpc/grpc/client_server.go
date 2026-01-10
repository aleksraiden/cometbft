package coregrpc

import (
	"context"
	"net"

	"google.golang.org/grpc"

	cmtnet "github.com/cometbft/cometbft/libs/net"
	"github.com/cometbft/cometbft/rpc/core"
	
	txstreampb "github.com/cometbft/cometbft/rpc/grpc"
	
	"github.com/cometbft/cometbft/rpc/grpc/txstreampb"
)

// Config is an gRPC server configuration.
//
// Deprecated: A new gRPC API will be introduced after v0.38.
type Config struct {
	MaxOpenConnections int
}

type txStreamServer struct {
    txstreampb.UnimplementedTxStreamServiceServer
    env *core.Environment
}

func NewTxStreamServer(env *core.Environment) txstreampb.TxStreamServiceServer {
    return &txStreamServer{env: env}
}

func (s *txStreamServer) BroadcastStream(stream txstreampb.TxStreamService_BroadcastStreamServer) error {
    ctx := stream.Context()
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }

        mode := strings.ToLower(req.GetMode())
        if mode == "" {
            mode = "sync"
        }

        // Вызов стандартной логики broadcast_tx_* ИЗНУТРИ
        var res *types.ResultBroadcastTx
        var respErr error

        switch mode {
        case "async":
            res, respErr = s.core.BroadcastTxAsync(ctx, req.GetTx())
        case "sync":
            res, respErr = s.core.BroadcastTxSync(ctx, req.GetTx())
        case "commit":
            resCommit, err := s.core.BroadcastTxCommit(ctx, req.GetTx())
            // конвертируй ResultBroadcastTxCommit -> ResultBroadcastTx
            respErr = err
        default:
            respErr = fmt.Errorf("unknown mode: %s", mode)
        }

        if respErr != nil {
            stream.Send(&txstreampb.TxStreamResponse{
                RequestId: req.GetRequestId(),
                Mode:      mode,
                Code:      1,
                Log:       respErr.Error(),
            })
            continue
        }

        stream.Send(&txstreampb.TxStreamResponse{
            RequestId: req.GetRequestId(),
            Mode:      mode,
            Code:      uint32(res.Code),
            Log:       res.Log,
            Data:      res.Data,
            TxHash:    fmt.Sprintf("%X", res.Hash),
        })
    }
}


// StartGRPCServer starts a new gRPC BroadcastAPIServer using the given
// net.Listener.
// NOTE: This function blocks - you may want to call it in a go-routine.
//
// Deprecated: A new gRPC API will be introduced after v0.38.
func StartGRPCServer(env *core.Environment, ln net.Listener) error {
	grpcServer := grpc.NewServer()
	RegisterBroadcastAPIServer(grpcServer, &broadcastAPI{env: env})
	
	txstreampb.RegisterTxStreamServiceServer(grpcServer, NewTxStreamServer(env))
	
	return grpcServer.Serve(ln)
}

// StartGRPCClient dials the gRPC server using protoAddr and returns a new
// BroadcastAPIClient.
//
// Deprecated: A new gRPC API will be introduced after v0.38.
func StartGRPCClient(protoAddr string) BroadcastAPIClient {
	conn, err := grpc.Dial(protoAddr, grpc.WithInsecure(), grpc.WithContextDialer(dialerFunc))
	if err != nil {
		panic(err)
	}
	return NewBroadcastAPIClient(conn)
}

func dialerFunc(_ context.Context, addr string) (net.Conn, error) {
	return cmtnet.Connect(addr)
}
