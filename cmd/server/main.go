package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/mctofu/music-library-grpc/go/mlibgrpc"
	"github.com/mctofu/musiclib/mlib"
	"github.com/mctofu/musiclib/mlib/filesys"
	"google.golang.org/grpc"
)

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
	s := grpc.NewServer()
	mlibgrpc.RegisterMusicLibraryServer(s,
		&server{
			library: &filesys.Library{
				RootPaths: []string{"/mnt/media/music/eac_flac_encoded", "/mnt/media/music/cindy", "/mnt/media/music/purchased"},
			},
		})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}

type server struct {
	mlibgrpc.UnimplementedMusicLibraryServer
	library *filesys.Library
}

// SayHello implements helloworld.GreeterServer
func (s *server) Browse(ctx context.Context, in *mlibgrpc.BrowseRequest) (*mlibgrpc.BrowseResponse, error) {
	log.Printf("Received: %v", in)

	browseOpts := mlib.BrowseOptions{TextFilter: strings.ToLower(in.Search)}
	items, err := s.library.Browse(ctx, in.Path, browseOpts)
	if err != nil {
		return nil, err
	}

	if in.Reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}

	return &mlibgrpc.BrowseResponse{
		Items: items,
	}, nil
}
