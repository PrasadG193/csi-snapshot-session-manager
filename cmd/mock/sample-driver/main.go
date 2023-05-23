package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"os"

	pb "github.com/PrasadG193/external-snapshot-session-access/pkg/grpc"
	grpcserver "github.com/PrasadG193/external-snapshot-session-access/pkg/grpc/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var csiAddress = flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket.")

//func main() {
//	flag.Parse()
//
//	log.Printf("listening at %s", *csiAddress)
//	listener, err := net.Listen("unix", *csiAddress)
//	if err != nil {
//		log.Fatalf("failed to listen: %v", err)
//	}
//
//	opts := []grpc.ServerOption{}
//	grpcServer := grpc.NewServer(opts...)
//	pb.RegisterVolumeSnapshotDeltaServiceServer(grpcServer, grpcserver.New())
//	if err := grpcServer.Serve(listener); err != nil {
//		log.Println(err)
//	}
//}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	cert := os.Getenv("CBT_SERVER_CERT")
	key := os.Getenv("CBT_SERVER_KEY")
	serverCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}
	s := grpc.NewServer(
		grpc.Creds(tlsCredentials),
	)
	reflection.Register(s)
	pb.RegisterVolumeSnapshotDeltaServiceServer(s, grpcserver.New())
	if err := s.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
