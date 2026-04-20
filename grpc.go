package EventBus

import (
	"encoding/json"
	"fmt"

	pb "github.com/ChrisEineke/EventBus/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// RpcServiceFireEvent is the RPC methods for firing events remotely.
	RpcServiceFireEvent = "Fire"
)

type GrpcWare struct {
	Handlerware
	registry Registry[*grpcRegistration]

	clients map[string]pb.EventServiceClient
	server  *grpcServer
}

func NewGrpcPeer() Handlerware {
	g := &GrpcWare{
		registry: Registry[*grpcRegistration]{},
		clients:  make(map[string]pb.EventServiceClient),
	}
	grpcServer := &grpcServer{GrpcWare: g}
	g.server = grpcServer
	return g
}

func (g *GrpcWare) OnUse(e *Event) error {
	return g.registry.Register(e, &grpcRegistration{
		eventPayload: make(chan []any),
	})
}

func (g *GrpcWare) OnDisuse(e *Event) error {
	_, err := g.registry.Deregister(e)
	return err
}

func (g *GrpcWare) OnPreFire(e *Event, args ...any) {
}

func (g *GrpcWare) OnPostFire(e *Event, args ...any) {
	registration, err := g.registry.Registration(e.N)
	if err != nil {
		return
	}
	registration.data.eventPayload <- args
}

func (g *GrpcWare) Connect(serverAddr string, opts ...grpc.DialOption) error {
	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return err
	}
	client := pb.NewEventServiceClient(conn)
	g.clients[serverAddr] = client
	return nil
}

type grpcServer struct {
	*GrpcWare

	pb.UnimplementedEventServiceServer
	pb.EventServiceServer
}

type grpcRegistration struct {
	eventPayload chan []any
}

func (g *grpcServer) Subscribe(subscription *pb.EventSubscription, stream grpc.ServerStreamingServer[pb.EventDelivery]) error {
	eventName := subscription.GetName()
	grpcRegistration, err := g.registry.Registration(eventName)
	if err != nil {
		return fmt.Errorf("event %s is not known", eventName)
	}
	for {
		var args []any = <-grpcRegistration.data.eventPayload
		var pbArgs []*anypb.Any = make([]*anypb.Any, len(args))
		for i, arg := range args {
			anyValue := &anypb.Any{}
			anyToPbAny(arg, anyValue)
			pbArgs[i] = anyValue
		}
		stream.Send(&pb.EventDelivery{Args: pbArgs})
	}
}

func anyToPbAny(inVal any, outVal *anypb.Any) error {
	bytes, err := json.Marshal(inVal)
	if err != nil {
		return err
	}
	bytesValue := &wrapperspb.BytesValue{Value: bytes}
	err = anypb.MarshalFrom(outVal, bytesValue, proto.MarshalOptions{})
	if err != nil {
		return err
	}
	return nil
}

func pbAnyToAny(inVal *anypb.Any, outVal *any) error {
	bytesValue := &wrappers.BytesValue{}
	err := anypb.UnmarshalTo(inVal, bytesValue, proto.UnmarshalOptions{})
	if err != nil {
		return err
	}
	return json.Unmarshal(bytesValue.Value, outVal)
}
