// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package walletv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// WalletServiceClient is the client API for WalletService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WalletServiceClient interface {
	UserCreate(ctx context.Context, in *UserCreateRequest, opts ...grpc.CallOption) (*UserCreateResponse, error)
	FundIn(ctx context.Context, in *FundInRequest, opts ...grpc.CallOption) (*FundInResponse, error)
	FundOut(ctx context.Context, in *FundOutRequest, opts ...grpc.CallOption) (*FundOutResponse, error)
}

type walletServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWalletServiceClient(cc grpc.ClientConnInterface) WalletServiceClient {
	return &walletServiceClient{cc}
}

func (c *walletServiceClient) UserCreate(ctx context.Context, in *UserCreateRequest, opts ...grpc.CallOption) (*UserCreateResponse, error) {
	out := new(UserCreateResponse)
	err := c.cc.Invoke(ctx, "/wallet.v1.WalletService/UserCreate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) FundIn(ctx context.Context, in *FundInRequest, opts ...grpc.CallOption) (*FundInResponse, error) {
	out := new(FundInResponse)
	err := c.cc.Invoke(ctx, "/wallet.v1.WalletService/FundIn", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) FundOut(ctx context.Context, in *FundOutRequest, opts ...grpc.CallOption) (*FundOutResponse, error) {
	out := new(FundOutResponse)
	err := c.cc.Invoke(ctx, "/wallet.v1.WalletService/FundOut", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WalletServiceServer is the server API for WalletService service.
// All implementations should embed UnimplementedWalletServiceServer
// for forward compatibility
type WalletServiceServer interface {
	UserCreate(context.Context, *UserCreateRequest) (*UserCreateResponse, error)
	FundIn(context.Context, *FundInRequest) (*FundInResponse, error)
	FundOut(context.Context, *FundOutRequest) (*FundOutResponse, error)
}

// UnimplementedWalletServiceServer should be embedded to have forward compatible implementations.
type UnimplementedWalletServiceServer struct {
}

func (UnimplementedWalletServiceServer) UserCreate(context.Context, *UserCreateRequest) (*UserCreateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserCreate not implemented")
}
func (UnimplementedWalletServiceServer) FundIn(context.Context, *FundInRequest) (*FundInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FundIn not implemented")
}
func (UnimplementedWalletServiceServer) FundOut(context.Context, *FundOutRequest) (*FundOutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FundOut not implemented")
}

// UnsafeWalletServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WalletServiceServer will
// result in compilation errors.
type UnsafeWalletServiceServer interface {
	mustEmbedUnimplementedWalletServiceServer()
}

func RegisterWalletServiceServer(s grpc.ServiceRegistrar, srv WalletServiceServer) {
	s.RegisterService(&WalletService_ServiceDesc, srv)
}

func _WalletService_UserCreate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).UserCreate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/wallet.v1.WalletService/UserCreate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).UserCreate(ctx, req.(*UserCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_FundIn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FundInRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).FundIn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/wallet.v1.WalletService/FundIn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).FundIn(ctx, req.(*FundInRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_FundOut_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FundOutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).FundOut(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/wallet.v1.WalletService/FundOut",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).FundOut(ctx, req.(*FundOutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WalletService_ServiceDesc is the grpc.ServiceDesc for WalletService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WalletService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "wallet.v1.WalletService",
	HandlerType: (*WalletServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UserCreate",
			Handler:    _WalletService_UserCreate_Handler,
		},
		{
			MethodName: "FundIn",
			Handler:    _WalletService_FundIn_Handler,
		},
		{
			MethodName: "FundOut",
			Handler:    _WalletService_FundOut_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "wallet/v1/wallet.proto",
}