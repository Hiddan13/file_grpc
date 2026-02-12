package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"file_grpc/api/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("address", "localhost:50051", "server address")
	action     = flag.String("action", "", "upload/download/list")
	filename   = flag.String("file", "", "file to upload or download")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewFileServiceClient(conn)

	switch *action {
	case "upload":
		if *filename == "" {
			log.Fatal("filename required for upload")
		}
		uploadFile(client, *filename)
	case "download":
		if *filename == "" {
			log.Fatal("filename required for download")
		}
		downloadFile(client, *filename)
	case "list":
		listFiles(client)
	default:
		log.Fatal("unknown action, use upload/download/list")
	}
}

func uploadFile(client pb.FileServiceClient, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.Upload(ctx)
	if err != nil {
		log.Fatalf("failed to start upload: %v", err)
	}

	// Отправляем частями по 64KB
	chunkSize := 64 * 1024
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		err := stream.Send(&pb.UploadRequest{
			Filename: filename,
			Chunk:    data[i:end],
		})
		if err != nil {
			log.Fatalf("failed to send chunk: %v", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}
	fmt.Printf("Uploaded: %s, size=%d bytes\n", resp.Message, resp.Size)
}

func downloadFile(client pb.FileServiceClient, filename string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.Download(ctx, &pb.DownloadRequest{Filename: filename})
	if err != nil {
		log.Fatalf("failed to start download: %v", err)
	}

	var data []byte
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to receive chunk: %v", err)
		}
		data = append(data, resp.Chunk...)
	}

	// Сохраняем как downloaded_<имя>
	outName := "downloaded_" + filename
	err = os.WriteFile(outName, data, 0644)
	if err != nil {
		log.Fatalf("failed to save file: %v", err)
	}
	fmt.Printf("Downloaded %s (%d bytes) to %s\n", filename, len(data), outName)
}

func listFiles(client pb.FileServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListFiles(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("failed to list files: %v", err)
	}

	fmt.Printf("%-20s | %-25s | %-25s | %s\n", "Filename", "Created At", "Updated At", "Size (bytes)")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, f := range resp.Files {
		fmt.Printf("%-20s | %-25s | %-25s | %d\n", f.Filename, f.CreatedAt, f.UpdatedAt, f.Size)
	}
}
