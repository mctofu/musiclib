package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/mctofu/music-library-grpc/go/mlibgrpc"
	"github.com/mctofu/musiclib/mlib"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error occurred: %v", err)
	}
}

func run() error {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	rootPathSetting, _ := os.LookupEnv("MUSICLIB_ROOT_PATHS")
	listenAddrSetting, _ := os.LookupEnv("MUSICLIB_LISTEN_ADDR")

	var rootPaths []string
	if rootPathSetting == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("MUSICLIB_ROOT_PATHS not set and could not detect home: %v", err)
		}
		rootPaths = append(rootPaths, path.Join(home, "Music"))
	} else {
		rootPaths = strings.Split(rootPathSetting, ",")
	}

	listenAddr := listenAddrSetting
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8337"
	}

	lis, err := net.Listen("tcp", listenAddr)
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

	go func() {
		defer signal.Stop(stop)
		<-stop
		log.Println("Stopping")
		s.GracefulStop()
		log.Println("Stopped")
	}()

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
	log.Printf("Received Browse: %v", in)
	startTime := time.Now()

	browseType, err := toMLibBrowseType(in.GetBrowseType())
	if err != nil {
		return nil, err
	}

	browseOpts := mlib.BrowseOptions{
		TextFilter: strings.ToLower(in.GetSearch()),
		BrowseType: browseType,
	}

	items, err := s.library.Browse(ctx, in.GetUri(), browseOpts)
	if err != nil {
		return nil, err
	}

	log.Printf("Browse: found %d items in %d ns", len(items), time.Since(startTime).Nanoseconds())

	if in.Reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}

	return &mlibgrpc.BrowseResponse{
		Items: toMLibGRPCItems(items),
	}, nil
}

func (s *server) Media(ctx context.Context, in *mlibgrpc.MediaRequest) (*mlibgrpc.MediaResponse, error) {
	log.Printf("Received Media: %v", in)
	startTime := time.Now()

	browseType, err := toMLibBrowseType(in.GetBrowseType())
	if err != nil {
		return nil, err
	}

	browseOpts := mlib.BrowseOptions{
		TextFilter: strings.ToLower(in.GetSearch()),
		BrowseType: browseType,
	}

	uris, err := s.library.Media(ctx, in.GetUri(), browseOpts)
	if err != nil {
		return nil, err
	}

	log.Printf("Media: found %d uris in %d ns", len(uris), time.Since(startTime).Nanoseconds())

	if in.Reverse {
		for i, j := 0, len(uris)-1; i < j; i, j = i+1, j-1 {
			uris[i], uris[j] = uris[j], uris[i]
		}
	}

	return &mlibgrpc.MediaResponse{
		Uris: uris,
	}, nil
}

func toMLibBrowseType(t mlibgrpc.BrowseType) (mlib.BrowseType, error) {
	switch t {
	case mlibgrpc.BrowseType_BROWSE_TYPE_ALBUM_ARTIST:
		return mlib.BrowseTypeAlbumArtist, nil
	case mlibgrpc.BrowseType_BROWSE_TYPE_FOLDER:
		return mlib.BrowseTypeFile, nil
	case mlibgrpc.BrowseType_BROWSE_TYPE_UNSPECIFIED:
		return mlib.BrowseTypeFile, nil
	default:
		return "", fmt.Errorf("unsupported browseType: %v", t)
	}
}

func toMLibGRPCItems(items []*mlib.BrowseItem) []*mlibgrpc.BrowseItem {
	result := make([]*mlibgrpc.BrowseItem, 0, len(items))
	for _, item := range items {
		result = append(result, &mlibgrpc.BrowseItem{
			Name:     item.Name,
			Uri:      item.URI,
			ImageUri: item.ImageURI,
			Folder:   item.Folder,
		})
	}

	return result
}
