package handler

import (
	"context"
	"grpcserver/api/gen/proto"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	// seed the random number generator
	_ = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type BakerHandler struct {
	proto.UnimplementedPancakeBakerServiceServer
	report *report
}

type report struct {
	sync.Mutex
	data map[proto.Pancake_Menu]int
}

func NewBakerHandler() *BakerHandler {
	return &BakerHandler{
		report: &report{
			data: make(map[proto.Pancake_Menu]int),
		},
	}
}

func (h *BakerHandler) Bake(ctx context.Context, req *proto.BakeRequest) (*proto.BakeResponse, error) {
	if req.Menu == proto.Pancake_UNKNOWN || req.Menu > proto.Pancake_PUMPKIN {
		return nil, status.Errorf(codes.InvalidArgument, "unknown pancake menu")
	}

	now := time.Now()
	h.report.Lock()
	h.report.data[req.Menu]++
	h.report.Unlock()

	return &proto.BakeResponse{
		Pancake: &proto.Pancake{
			Menu:           req.Menu,
			ChefName:       "gami",
			TechnicalScore: rand.Float32(),
			CreateTime: &timestamp.Timestamp{
				Seconds: now.Unix(),
				Nanos:   int32(now.Nanosecond()),
			},
		},
	}, nil
}

func (h *BakerHandler) Report(ctx context.Context, req *proto.ReportRequest) (*proto.ReportResponse, error) {
	counts := make([]*proto.Report_BakeCount, 0)

	h.report.Lock()
	for k, v := range h.report.data {
		counts = append(counts, &proto.Report_BakeCount{
			Menu:  k,
			Count: int32(v),
		})
	}
	h.report.Unlock()

	return &proto.ReportResponse{
		Report: &proto.Report{
			BakeCounts: counts,
		},
	}, nil
}
