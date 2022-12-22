package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"

	"testGRPC/data"

	"testGRPC/gw/common"
	"testGRPC/gw/entity"
	"testGRPC/gw/service"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	logrus "github.com/sirupsen/logrus"

	pbMember "github.com/beautiful-store/grpc/test/member"
	pbStream "github.com/beautiful-store/grpc/test/stream"
	pb "github.com/beautiful-store/grpc/test/user"
)

const portNumber = "9000"

type userServer struct {
	pb.UserServer
}

type memberServer struct {
	pbMember.MemberSeviceServer
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userID := req.UserId

	var userMessage *pb.UserMessage
	for _, u := range data.Users {
		if u.UserId != userID {
			continue
		}
		userMessage = u
		break
	}

	return &pb.GetUserResponse{
		UserMessage: userMessage,
	}, nil
}

// ListUsers returns all user messages
func (s *userServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	fmt.Println("ListUsers token:", ctx.Value("token"))

	userMessages := make([]*pb.UserMessage, len(data.Users))
	for i, u := range data.Users {
		userMessages[i] = u
	}

	return &pb.ListUsersResponse{
		UserMessages: userMessages,
	}, nil
}

type streamServer struct {
	pbStream.StreamServiceServer
}

func (s *streamServer) GetURL(in *pbStream.GetURLRequest, srv pbStream.StreamService_GetURLServer) error {
	url := in.Url

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// b, err := ioutil.ReadAll(resp.Body)
	// r := pbStream.GetURLResponse{Content: b}
	// if err := srv.Send(&r); err != nil {
	// 	log.Printf("send error %v", err)
	// }

	p := make([]byte, 500*1024)
	for {
		// read `p` bytes from `src`
		n, err := resp.Body.Read(p)

		// handle error
		if err == io.EOF {
			fmt.Println("--end-of-file--")
			break
		} else if err != nil {
			fmt.Println("Oops! Some error occured!", err)
			break
		}

		r := pbStream.GetURLResponse{Content: p[:n]}
		if err := srv.Send(&r); err != nil {
			log.Printf("send error %v", err)
		}
	}

	return nil
}

func (s *memberServer) GetList(ctx context.Context, req *pbMember.IDs) (*pbMember.MemberList, error) {
	logrus.Trace("")

	ids, err := service.ValidateRequesst(req.Ids)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	list, err := service.GetList(ctx, ids)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	// m := pbMember.Member{
	// 	MemberID:     1,
	// 	MemberName:   "MemberName",
	// 	Mobile:       "Mobile",
	// 	MaskedMobile: "MaskedMobile",
	// 	Email:        "Email",
	// 	OrgID:        1,
	// }

	// mList.MemberList = append(mList.MemberList, &m)
	mList := ConverToMemberServiceProto(list)

	return &mList, nil
}

func ConverToMemberServiceProto(list []entity.Member) pbMember.MemberList {
	mList := pbMember.MemberList{}
	for _, l := range list {
		m := pbMember.Member{
			MemberID:     l.MemberID,
			MemberName:   l.MemberName,
			Mobile:       l.Mobile,
			MaskedMobile: l.MaskedMobile,
			Email:        l.Email,
			OrgID:        l.OrgID,
		}
		mList.MemberList = append(mList.MemberList, &m)
	}

	return mList
}

func main() {
	xormDB := common.ConfigureDatabase()

	lis, err := net.Listen("tcp", ":"+portNumber)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// logrus.ErrorKey = "grpc.error"
	logrusEntry := logrus.NewEntry(logrus.StandardLogger())

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			customMiddleware(),
			customMiddlewareConnectDB(xormDB),
			// grpc_auth.UnaryServerInterceptor(customAuthFunc),
			// 		grpc_ctxtags.UnaryServerInterceptor(),
			// 		grpc_opentracing.UnaryServerInterceptor(),
			// 		grpc_prometheus.UnaryServerInterceptor,
			// 		grpc_zap.UnaryServerInterceptor(zapLogger),
			// 		grpc_auth.UnaryServerInterceptor(myAuthFunction),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
			grpc_recovery.UnaryServerInterceptor(),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			customMiddleware2(),
			// grpc_ctxtags.StreamServerInterceptor(),
			// grpc_opentracing.StreamServerInterceptor(),
			// grpc_prometheus.StreamServerInterceptor,
			// grpc_zap.StreamServerInterceptor(zapLogger),
			// grpc_auth.StreamServerInterceptor(customAuthFunc),
			// grpc_logrus.StreamServerInterceptor(logrusEntry),
			grpc_recovery.StreamServerInterceptor(),
		)),
	)

	pb.RegisterUserServer(grpcServer, &userServer{})
	pbStream.RegisterStreamServiceServer(grpcServer, &streamServer{})
	pbMember.RegisterMemberSeviceServer(grpcServer, &memberServer{})

	log.Printf("start gRPC server on %s port", portNumber)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func customMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		// log.Print("(Unary) Requested at:", time.Now())

		resp, err := handler(ctx, req)
		return resp, err
	}
}

func customMiddleware2() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// log.Print("(Stream) Requested at:", time.Now())

		wrapped := grpc_middleware.WrapServerStream(stream)

		return handler(srv, wrapped)
	}
}

func customMiddlewareConnectDB(xormDB *xorm.Engine) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		log.Print("customMiddlewareConnectDB:", time.Now())

		newCtx := context.WithValue(ctx, "xormDB", xormDB)

		resp, err := handler(newCtx, req)
		return resp, err
	}
}

// func ConfigureDatabase() *xorm.Engine {
// 	uid := "root"
// 	pwd := os.Getenv("SHARING_PLATFORM_DB_PASSWORD")

// 	dbConnection := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/sharing?charset=utf8mb4&collation=utf8mb4_unicode_ci", uid, pwd)
// 	xormDb, err := xorm.NewEngine("mysql", dbConnection)
// 	if err != nil {
// 		panic(fmt.Errorf("Database open error: " + err.Error()))
// 	}

// 	return xormDb
// }

func customAuthFunc(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}
	// fmt.Println("Token:", token)

	newCtx := context.WithValue(ctx, "token", token)

	return newCtx, nil
}
