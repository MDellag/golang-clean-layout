package grpc

import (
	"clean-arq-layout/internal/domain/dto/request"
	"clean-arq-layout/internal/domain/interfaces"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

type UserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserGrpcService struct {
	userService interfaces.UserService
}

func NewUserGrpcService(userService interfaces.UserService) *UserGrpcService {
	return &UserGrpcService{
		userService: userService,
	}
}

func (s *UserGrpcService) CreateUser(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	createDTO := request.CreateUserDTO{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}

	userDTO, err := s.userService.CreateUser(ctx, createDTO)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &UserResponse{
		ID:    userDTO.ID,
		Email: userDTO.Email,
		Name:  userDTO.Name,
	}, nil
}

func (s *UserGrpcService) GetUser(ctx context.Context, id string) (*UserResponse, error) {
	userDTO, err := s.userService.GetUserByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &UserResponse{
		ID:    userDTO.ID,
		Email: userDTO.Email,
		Name:  userDTO.Name,
	}, nil
}

type GrpcServer struct {
	userService interfaces.UserService
	grpcServer  *grpc.Server
	port        string
}

func NewGrpcServer(userService interfaces.UserService, port string) *GrpcServer {
	return &GrpcServer{
		userService: userService,
		grpcServer:  grpc.NewServer(),
		port:        port,
	}
}

func (s *GrpcServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	userGrpcService := NewUserGrpcService(s.userService)

	s.grpcServer.RegisterService(&grpc.ServiceDesc{
		ServiceName: "UserService",
		HandlerType: (*UserGrpcService)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CreateUser",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(UserRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					if interceptor == nil {
						return srv.(*UserGrpcService).CreateUser(ctx, in)
					}
					info := &grpc.UnaryServerInfo{
						Server:     srv,
						FullMethod: "/UserService/CreateUser",
					}
					handler := func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(*UserGrpcService).CreateUser(ctx, req.(*UserRequest))
					}
					return interceptor(ctx, in, info, handler)
				},
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "user_service.proto",
	}, userGrpcService)

	log.Printf("gRPC server starting on port %s", s.port)
	return s.grpcServer.Serve(listener)
}

func (s *GrpcServer) Stop() {
	log.Println("Shutting down gRPC server...")
	s.grpcServer.GracefulStop()
}