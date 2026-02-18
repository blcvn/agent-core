package main

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	pb "github.com/blcvn/kratos-proto/go/ba-agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Connect to BA Agent Service
	conn, err := grpc.Dial("localhost:9088", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewBAAgentServiceClient(conn)

	// 2. Read Input File
	inputBytes, err := ioutil.ReadFile("input_complex.txt")
	if err != nil {
		log.Fatalf("failed to read input file: %v", err)
	}
	content := string(inputBytes)

	log.Printf("Sending ExecuteTask request with content length: %d...", len(content))

	// 3. Call ExecuteTask
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req := &pb.ExecuteTaskRequest{
		Payload: &pb.ExecuteTaskPayload{
			TaskDescription: content, // Fallback used in pipeline.go
			WorkflowMode:    "pipeline",
			// ActionParams: map[string]string{"content": content}, // Proto issue preventing this
		},
	}

	start := time.Now()
	resp, err := client.ExecuteTask(ctx, req)
	if err != nil {
		log.Fatalf("ExecuteTask failed: %v", err)
	}
	duration := time.Since(start)

	log.Printf("Request completed in %v", duration)
	log.Printf("Result Code: %s", resp.Result.Code)
	log.Printf("Result Message: %s", resp.Result.Message)

	if resp.Task != nil {
		log.Printf("Task Status: %s", resp.Task.Status)

		// 4. Save Output
		if resp.Task.FinalResponse != "" {
			outputFile := "output_pipeline.md"
			err = ioutil.WriteFile(outputFile, []byte(resp.Task.FinalResponse), 0644)
			if err != nil {
				log.Printf("Failed to write output file: %v", err)
			} else {
				log.Printf("Successfully wrote output to %s", outputFile)
			}
		} else {
			log.Println("Warning: FinalResponse is empty")
		}
	}
}
