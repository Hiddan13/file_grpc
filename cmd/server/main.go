package main

import (
	"log"
	"net"

	"github.com/Hiddan13/file_grpc/api/proto/pb"
	"github.com/Hiddan13/file_grpc/internal/config"
	"github.com/Hiddan13/file_grpc/internal/repository"
	"github.com/Hiddan13/file_grpc/internal/service"
	grpcTransport "github.com/Hiddan13/file_grpc/internal/transport/grpc"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	// init репозитория
	repo, err := repository.NewFilesRepository(cfg.StoragePath)
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}

	// init сервиса
	fileservice := service.NewFileService(repo)

	// Создаём gRPC сервер
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	// Регистрация обработчиков
	fileServer := grpcTransport.NewFileServer(fileservice, cfg.UploadLimit, cfg.DownloadLimit, cfg.ListLimit)
	pb.RegisterFileServiceServer(grpcServer, fileServer)

	log.Printf("gRPC server listening on %s", cfg.GRPCPort)
	log.Printf("storage path: %s", cfg.StoragePath)
	log.Printf("limits: upload=%d, download=%d, list=%d", cfg.UploadLimit, cfg.DownloadLimit, cfg.ListLimit)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
