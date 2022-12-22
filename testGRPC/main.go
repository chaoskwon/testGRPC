package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	pbMember "github.com/beautiful-store/grpc/test/member"
	pbStream "github.com/beautiful-store/grpc/test/stream"
	pb "github.com/beautiful-store/grpc/test/user"
	"github.com/labstack/echo/v4"

	controller "testGRPC/gw/controller"
	echomiddleware "testGRPC/gw/echomiddleware"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var gRPCServerPortNumber = "9000"

func main() {
	e := echo.New()
	echomiddleware.ConfigueEcho(e)

	gwmux := runtime.NewServeMux()

	err := pb.RegisterUserHandlerFromEndpoint(context.Background(), gwmux, "localhost:"+gRPCServerPortNumber, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		panic(err)
	}
	err = pbStream.RegisterStreamServiceHandlerFromEndpoint(context.Background(), gwmux, "localhost:"+gRPCServerPortNumber, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		panic(err)
	}
	err = pbMember.RegisterMemberSeviceHandlerFromEndpoint(context.Background(), gwmux, "localhost:"+gRPCServerPortNumber, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		panic(err)
	}

	e.GET("/members/org/list", func(c echo.Context) error {
		list, err := controller.MemberController(c)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, list)
	})

	if err := http.ListenAndServe(":7000", allHandler(gwmux, e)); err != nil {
		panic(err)
	}
}

func allHandler(mux *runtime.ServeMux, e *echo.Echo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/grpc") {
			fmt.Println("==grpc")
			mux.ServeHTTP(w, r)
		} else {
			fmt.Println("==echo")
			e.ServeHTTP(w, r)
		}
	})
}
