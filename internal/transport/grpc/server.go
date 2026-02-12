package grpc

import (
	"context"
	"io"
	"log"
	"time"

	"file_grpc/api/proto/pb"
	"file_grpc/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileServer struct {
	pb.UnimplementedFileServiceServer
	fileService       *service.FileService
	uploadSemophore   chan struct{}
	downloadSemophore chan struct{}
	listSemophore     chan struct{}
}

func NewFileServer(fileService *service.FileService, uploadLimit, downloadLimit, listLimit int) *FileServer {
	return &FileServer{
		fileService:       fileService,
		uploadSemophore:   make(chan struct{}, uploadLimit),
		downloadSemophore: make(chan struct{}, downloadLimit),
		listSemophore:     make(chan struct{}, listLimit),
	}
}

// Загрузка файла на ссервер стрим
func (s *FileServer) Upload(stream pb.FileService_UploadServer) error {
	// Лимит
	select {
	case s.uploadSemophore <- struct{}{}:
		defer func() {
			<-s.uploadSemophore
			log.Printf("[UPLOAD] завершён, активных загрузок: %d", len(s.uploadSemophore))
		}()
	default:
		log.Printf("[UPLOAD] ОТКАЗ: превышен лимит (%d)", cap(s.uploadSemophore))
		return status.Error(codes.ResourceExhausted, "upload limit exceeded")
	}

	req, err := stream.Recv()
	if err != nil {
		log.Printf("[UPLOAD] ошибка получения первого чанка: %v", err)
		return err
	}
	filename := req.GetFilename()
	var totalSize int
	var chunkCount = 1
	var data []byte
	data = append(data, req.GetChunk()...)
	totalSize += len(req.GetChunk())

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[UPLOAD] ошибка получения чанка: %v", err)
			return err
		}
		chunkCount++
		data = append(data, req.GetChunk()...)
		totalSize += len(req.GetChunk())
	}

	log.Printf("[UPLOAD] файл=%s, чанков=%d, размер=%d байт", filename, chunkCount, totalSize)

	if err := s.fileService.SaveFile(stream.Context(), filename, data); err != nil {
		log.Printf("[UPLOAD] ошибка сохранения: %v", err)
		return status.Errorf(codes.Internal, "failed to save file: %v", err)
	}

	log.Printf("[UPLOAD] успешно сохранён: %s", filename)
	return stream.SendAndClose(&pb.UploadResponse{
		Message: "file uploaded successfully",
		Size:    int64(totalSize),
	})
}

// Скачаем файл через стрим
func (s *FileServer) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {
	select {
	case s.downloadSemophore <- struct{}{}:
		defer func() {
			<-s.downloadSemophore
			log.Printf("[DOWNLOAD] завершён, активных скачиваний: %d", len(s.downloadSemophore))
		}()
	default:
		log.Printf("[DOWNLOAD] ОТКАЗ: превышен лимит (%d)", cap(s.downloadSemophore))
		return status.Error(codes.ResourceExhausted, "download limit exceeded")
	}

	filename := req.GetFilename()
	log.Printf("[DOWNLOAD] запрос файла: %s", filename)

	data, err := s.fileService.GetFile(stream.Context(), filename)
	if err != nil {
		log.Printf("[DOWNLOAD] файл не найден: %s", filename)
		return status.Errorf(codes.NotFound, "file not found: %v", err)
	}

	_ = s.fileService.UpdateAccess(stream.Context(), filename)

	const chunkSize = 64 * 1024
	chunks := 0
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		if err := stream.Send(&pb.DownloadResponse{Chunk: data[i:end]}); err != nil {
			log.Printf("[DOWNLOAD] ошибка отправки чанка: %v", err)
			return err
		}
		chunks++
	}
	log.Printf("[DOWNLOAD] отправлен файл=%s, размер=%d, чанков=%d", filename, len(data), chunks)
	return nil
}

// Получаем список файлов
func (s *FileServer) ListFiles(ctx context.Context, _ *pb.Empty) (*pb.ListFilesResponse, error) {
	// 1. Лимит
	select {
	case s.listSemophore <- struct{}{}:
		defer func() {
			<-s.listSemophore
			log.Printf("[LIST] завершён, активных запросов: %d", len(s.listSemophore))
		}()
	default:
		log.Printf("[LIST] ОТКАЗ: превышен лимит (%d)", cap(s.listSemophore))
		return nil, status.Error(codes.ResourceExhausted, "list limit exceeded")
	}

	// Получаем список
	metas, err := s.fileService.ListFiles(ctx)
	if err != nil {
		log.Printf("[LIST] ошибка получения списка: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list files: %v", err)
	}
	log.Printf("[LIST] найдено файлов: %d", len(metas))
	// Преобразование в pb
	pbFiles := make([]*pb.FileInfo, 0, len(metas))
	for _, m := range metas {
		pbFiles = append(pbFiles, &pb.FileInfo{
			Filename:  m.Filename,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
			UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
			Size:      m.Size,
		})
	}
	return &pb.ListFilesResponse{Files: pbFiles}, nil
}
