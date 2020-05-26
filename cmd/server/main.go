package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/mctofu/music-library-grpc/go/mlibgrpc"
	"github.com/mctofu/musiclib/mlib"
	"google.golang.org/grpc"
)

var rootPaths = []string{"/mnt/media/music/eac_flac_encoded", "/mnt/media/music/cindy", "/mnt/media/music/purchased"}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error occurred: %v", err)
	}
}

func run() error {
	lis, err := net.Listen("tcp", "127.0.0.1:8337")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Println("Loading library")
	library, err := mlib.NewLibrary(context.Background(), rootPaths)
	if err != nil {
		return fmt.Errorf("failed to init library: %v", err)
	}
	log.Println("Loaded library")

	s := grpc.NewServer()
	mlibgrpc.RegisterMusicLibraryServer(s,
		&server{
			library: library,
		})

	log.Println("Starting server")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}

type server struct {
	mlibgrpc.UnimplementedMusicLibraryServer
	library *mlib.Library
}

func (s *server) Browse(ctx context.Context, in *mlibgrpc.BrowseRequest) (*mlibgrpc.BrowseResponse, error) {
	log.Printf("Received: %v", in)
	startTime := time.Now()

	browseOpts := mlib.BrowseOptions{TextFilter: strings.ToLower(in.Search)}
	items, err := s.library.Browse(ctx, in.Path, browseOpts)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d items in %d ns", len(items), time.Since(startTime).Nanoseconds())

	if in.Reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}

	return &mlibgrpc.BrowseResponse{
		Items: toMLibGRPCItems(items),
	}, nil
}

func toMLibGRPCItems(items []*mlib.BrowseItem) []*mlibgrpc.BrowseItem {
	result := make([]*mlibgrpc.BrowseItem, 0, len(items))
	for _, item := range items {
		result = append(result, &mlibgrpc.BrowseItem{
			Name:   item.Name,
			Uri:    item.URI,
			Folder: item.Folder,
		})
	}

	return result
}
