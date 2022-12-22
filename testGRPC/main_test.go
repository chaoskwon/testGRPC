package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testGRPC/gw/common"
	"testGRPC/gw/controller"
	"testing"

	pbMember "github.com/beautiful-store/grpc/test/member"
	"github.com/go-xorm/xorm"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var xormDB *xorm.Engine

func setupDB() {
	if xormDB == nil {
		xormDB = common.ConfigureDatabase()
	}
}

func runMux() *runtime.ServeMux {
	var gRPCServerPortNumber = "9000"

	mux := runtime.NewServeMux()
	err := pbMember.RegisterMemberSeviceHandlerFromEndpoint(context.Background(), mux, "localhost:"+gRPCServerPortNumber, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		fmt.Println("##", err.Error())
		panic(err)
	}
	return mux
}

func runConn() *grpc.ClientConn {
	conn, _ := grpc.Dial("localhost:9000",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	return conn
}

func TestAll(t *testing.T) {
	mux := runMux()
	conn := runConn()

	idCount := 1
	for i := 0; i < 3; i++ {
		idCount *= 10

		t.Run(fmt.Sprintf("group %d count", idCount), func(t *testing.T) {
			t.Run("**grpcAPI", func(t *testing.T) {
				RunGrpcApi(mux, idCount)
			})
			t.Run("**grpc", func(t *testing.T) {
				RunGrpc(conn, idCount)
			})
		})
	}
}

func TestGrpcApi(t *testing.T) {
	mux := runMux()

	idCount := 10
	t.Run("TestGrpcApi", func(t *testing.T) {
		RunGrpcApi(mux, idCount)
	})

	// idCount := 1
	// for i := 0; i < 3; i++ {
	// 	idCount *= 10

	// 	t.Run(fmt.Sprintf("mux %d count", idCount), func(t *testing.T) {
	// 		RunGrpcApi(mux, idCount)
	// 	})
	// }
}

func TestGrpc(t *testing.T) {
	conn := runConn()

	idCount := 10
	t.Run("grpc", func(t *testing.T) {
		RunGrpc(conn, idCount)
	})

	// for i := 0; i < 3; i++ {
	// 	idCount *= 10

	// 	t.Run(fmt.Sprintf("grpc %d count", idCount), func(t *testing.T) {
	// 		RunGrpc(conn, idCount)
	// 	})
	// }
}

// func TestApi(t *testing.T) {
// 	e := echo.New()

// 	idCount := 1
// 	for i := 0; i < 3; i++ {
// 		idCount *= 10

// 		t.Run(fmt.Sprintf("api %d count", idCount), func(t *testing.T) {
// 			RunApi(e, idCount)
// 		})
// 	}
// }

func BenchmarkGrpcApi(b *testing.B) {
	mux := runMux()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunGrpcApi(mux, 10)
	}
}

func BenchmarkGrpc(b *testing.B) {
	conn := runConn()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunGrpc(conn, 10)
	}
}

func BenchmarkApi(b *testing.B) {
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunApi(e, 10)
	}
}

// func BenchmarkGrpcApiParallel(b *testing.B) {
// 	mux := runMux()

// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			RunGrpcApi(mux, 10)
// 		}
// 	})
// }

// func BenchmarkApiParallel(b *testing.B) {
// 	e := echo.New()

// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			RunApi(e, 10)
// 		}
// 	})
// }

// func BenchmarkGrpcParallel(b *testing.B) {
// 	conn := runConn()

// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			RunGrpc(conn, 10)
// 		}
// 	})
// }

func RunGrpc(conn *grpc.ClientConn, count int) {
	member := pbMember.NewMemberSeviceClient(conn)
	ids := &pbMember.IDs{Ids: GetIDList(count)}
	list, err := member.GetList(context.Background(), ids)
	if err != nil {
		fmt.Printf("error:%+v", list)
		return
	}

	// fmt.Println("*Member Count", len(list.MemberList))
}

func GetIDList(max int) string {
	ids := "1"
	for i := 0; i < max; i++ {
		ids += fmt.Sprintf(",%d", rand.Intn(115291))
	}

	return ids
}

func RunApi(e *echo.Echo, count int) {
	path := fmt.Sprintf("/members/org/list?ids=%s", GetIDList(count))
	r := httptest.NewRequest(http.MethodGet, path, nil)

	setupDB()

	rec := httptest.NewRecorder()
	c := e.NewContext(r, rec)
	req := r.WithContext(context.WithValue(c.Request().Context(), "xormDB", xormDB))
	c.SetRequest(req)

	l, err := controller.MemberController(c)
	if err != nil {
		fmt.Printf("error : %v : %v", err, l)
	}

	// fmt.Println("****", len(l))
}

func RunGrpcApi(mux *runtime.ServeMux, count int) {
	path := fmt.Sprintf("/grpc/members/org/list?ids=%s", GetIDList(count))
	r := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		fmt.Printf("error(status) : %v", rr)
		return
	}

	// m := make(map[string]interface{})
	// err := json.Unmarshal(rr.Body.Bytes(), &m)
	// if err != nil {
	// 	fmt.Printf("error(unmarsharl): %v", err)
	// 	return
	// }

	// results, ok := m["memberList"].([]interface{})
	// if !ok {
	// 	fmt.Printf("error(convert):%v", m["memberList"])
	// 	return
	// }
	// fmt.Println("*Member Count", len(results))
}
