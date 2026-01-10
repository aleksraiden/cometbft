package grpc

import (
    "context"
    "net"
    "testing"
    "time"

    "github.com/cometbft/cometbft/rpc/core"
    "github.com/cometbft/cometbft/rpc/grpc/txstreampb"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func TestTxStreamServer(t *testing.T) {
    // 1. Создаём mock Environment для broadcast_tx_*
    mockEnv := &mockCoreEnvironment{
        // минимальный мок: возвращает валидный ответ для sync
        broadcastSyncFn: func(ctx context.Context, tx []byte) (*types.ResultBroadcastTx, error) {
            return &types.ResultBroadcastTx{
                Code:   0,
                Log:    "OK",
                Hash:   []byte("deadbeef"),
                Data:   []byte("data"),
            }, nil
        },
    }

    // 2. Создаём in-memory listener
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer ln.Close()

    // 3. Запускаем gRPC сервер в горутине
    grpcServer := grpc.NewServer()
    txstreampb.RegisterTxStreamServiceServer(grpcServer, NewTxStreamServer(mockEnv))
    done := make(chan struct{})
    go func() {
        defer close(done)
        grpcServer.Serve(ln)
    }()
    defer grpcServer.GracefulStop(context.Background())

    // 4. Ждём готовности
    conn, err := grpc.Dial(ln.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
    require.NoError(t, err)
    defer conn.Close()

    client := txstreampb.NewTxStreamServiceClient(conn)

    // 5. Тестируем стрим
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    stream, err := client.BroadcastStream(ctx)
    require.NoError(t, err)

    // Отправляем тестовую транзакцию
    req := &txstreampb.TxStreamRequest{
        Tx:       []byte("fake_tx_bytes"),
        Mode:     "sync",
        RequestId: "test-123",
    }
    err = stream.Send(req)
    require.NoError(t, err)

    // Читаем ответ
    resp, err := stream.Recv()
    require.NoError(t, err)
    assert.Equal(t, "test-123", resp.RequestId)
    assert.Equal(t, uint32(0), resp.Code)
    assert.Equal(t, "OK", resp.Log)
    assert.Equal(t, "DEADBEEF", resp.TxHash) // hex uppercase
    assert.Equal(t, []byte("data"), resp.Data)

    // Закрываем стрим
    err = stream.CloseSend()
    require.NoError(t, err)
}

// минимальный мок для core.Environment
type mockCoreEnvironment struct {
    broadcastSyncFn func(ctx context.Context, tx []byte) (*types.ResultBroadcastTx, error)
}

func (m *mockCoreEnvironment) BroadcastTxSync(ctx context.Context, tx []byte) (*types.ResultBroadcastTx, error) {
    return m.broadcastSyncFn(ctx, tx)
}

func (m *mockCoreEnvironment) BroadcastTxAsync(ctx context.Context, tx []byte) (*types.ResultBroadcastTx, error) {
    return &types.ResultBroadcastTx{Code: 0}, nil
}

func (m *mockCoreEnvironment) BroadcastTxCommit(ctx context.Context, tx []byte) (*types.ResultBroadcastTxCommit, error) {
    return &types.ResultBroadcastTxCommit{Code: 0}, nil
}

// остальные методы Environment возвращают заглушки
func (m *mockCoreEnvironment) Status(ctx context.Context) (*coretypes.ResultStatus, error) {
    return nil, nil
}

Вот дополнительные тесты для полного покрытия твоего gRPC‑стрима. Добавь их в тот же txstream_test.go:

go
func TestTxStream_InvalidMode(t *testing.T) {
    mockEnv := &mockCoreEnvironment{
        broadcastSyncFn: func(ctx context.Context, tx []byte) (*types.ResultBroadcastTx, error) {
            return nil, fmt.Errorf("should not be called")
        },
    }

    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer ln.Close()

    grpcServer := grpc.NewServer()
    txstreampb.RegisterTxStreamServiceServer(grpcServer, NewTxStreamServer(mockEnv))
    done := make(chan struct{})
    go func() {
        defer close(done)
        grpcServer.Serve(ln)
    }()
    defer grpcServer.GracefulStop(context.Background())

    conn, err := grpc.Dial(ln.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
    require.NoError(t, err)
    defer conn.Close()

    client := txstreampb.NewTxStreamServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    stream, err := client.BroadcastStream(ctx)
    require.NoError(t, err)

    // Отправляем запрос с неизвестным mode
    req := &txstreampb.TxStreamRequest{
        Tx:       []byte("fake_tx"),
        Mode:     "invalid_mode",
        RequestId: "test-invalid-1",
    }
    err = stream.Send(req)
    require.NoError(t, err)

    resp, err := stream.Recv()
    require.NoError(t, err)

    // Проверяем ошибку
    assert.Equal(t, "test-invalid-1", resp.RequestId)
    assert.Equal(t, uint32(1), resp.Code)  // code=1 для ошибки
    assert.Contains(t, resp.Log, "unknown mode: invalid_mode")
}

func TestTxStream_ErrorResponse(t *testing.T) {
    mockEnv := &mockCoreEnvironment{
        // мок возвращает ошибку для sync
        broadcastSyncFn: func(ctx context.Context, tx []byte) (*types.ResultBroadcastTx, error) {
            return nil, fmt.Errorf("mock broadcast error")
        },
    }

    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer ln.Close()

    grpcServer := grpc.NewServer()
    txstreampb.RegisterTxStreamServiceServer(grpcServer, NewTxStreamServer(mockEnv))
    done := make(chan struct{})
    go func() {
        defer close(done)
        grpcServer.Serve(ln)
    }()
    defer grpcServer.GracefulStop(context.Background())

    conn, err := grpc.Dial(ln.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
    require.NoError(t, err)
    defer conn.Close()

    client := txstreampb.NewTxStreamServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    stream, err := client.BroadcastStream(ctx)
    require.NoError(t, err)

    req := &txstreampb.TxStreamRequest{
        Tx:       []byte("error_tx"),
        Mode:     "sync",  // вызовет broadcastSyncFn
        RequestId: "test-error-1",
    }
    err = stream.Send(req)
    require.NoError(t, err)

    resp, err := stream.Recv()
    require.NoError(t, err)

    // Проверяем передачу ошибки из broadcast_tx_*
    assert.Equal(t, "test-error-1", resp.RequestId)
    assert.Equal(t, uint32(1), resp.Code)
    assert.Contains(t, resp.Log, "mock broadcast error")
}

func TestTxStream_MultipleRequests(t *testing.T) {
    mockEnv := &mockCoreEnvironment{
        broadcastSyncFn: func(ctx context.Context, tx []byte) (*types.ResultBroadcastTx, error) {
            return &types.ResultBroadcastTx{
                Code: 0,
                Log:  "OK",
                Hash: tx[:4], // простая заглушка
            }, nil
        },
    }

    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer ln.Close()

    grpcServer := grpc.NewServer()
    txstreampb.RegisterTxStreamServiceServer(grpcServer, NewTxStreamServer(mockEnv))
    done := make(chan struct{})
    go func() {
        defer close(done)
        grpcServer.Serve(ln)
    }()
    defer grpcServer.GracefulStop(context.Background())

    conn, err := grpc.Dial(ln.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
    require.NoError(t, err)
    defer conn.Close()

    client := txstreampb.NewTxStreamServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    stream, err := client.BroadcastStream(ctx)
    require.NoError(t, err)

    // Отправляем 3 запроса подряд
    requests := []*txstreampb.TxStreamRequest{
        {Tx: []byte("tx1"), Mode: "sync", RequestId: "r1"},
        {Tx: []byte("tx2"), Mode: "async", RequestId: "r2"},
        {Tx: []byte("tx3"), Mode: "sync", RequestId: "r3"},
    }

    for _, req := range requests {
        err := stream.Send(req)
        require.NoError(t, err)
    }

    // Читаем 3 ответа
    responses := make(map[string]*txstreampb.TxStreamResponse)
    for i := 0; i < 3; i++ {
        resp, err := stream.Recv()
        require.NoError(t, err)
        responses[resp.RequestId] = resp
    }

    // Проверяем все ответы
    for _, req := range requests {
        resp := responses[req.RequestId]
        assert.Equal(t, uint32(0), resp.Code)
        assert.Equal(t, req.Mode, resp.Mode)
        assert.Equal(t, 6, len(resp.TxHash)) // hex из первых 4 байт tx
    }
}

