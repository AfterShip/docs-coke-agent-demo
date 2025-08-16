package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type FileSystemTool struct{}

type ListFilesRequest struct {
	Directory string `json:"directory" description:"The directory path to list files from"`
}

type ListFilesResponse struct {
	Files []string `json:"files" description:"List of file names in the directory"`
	Error string   `json:"error,omitempty" description:"Error message if operation failed"`
}

type WriteFileRequest struct {
	FilePath string `json:"file_path" description:"The full path to the file to write to"`
	Content  string `json:"content" description:"The content to write to the file"`
}

type WriteFileResponse struct {
	Success bool   `json:"success" description:"Whether the write operation was successful"`
	Error   string `json:"error,omitempty" description:"Error message if operation failed"`
}

type ReadFileRequest struct {
	FilePath string `json:"file_path" description:"The full path to the file to read"`
}

type ReadFileResponse struct {
	Content string `json:"content" description:"The content of the file"`
	Error   string `json:"error,omitempty" description:"Error message if operation failed"`
}

func newFileSystemTool() []tool.BaseTool {
	fsTool := &FileSystemTool{}

	listTool, _ := utils.InferTool[ListFilesRequest, ListFilesResponse](
		"list_files",
		"List files in a specific directory",
		fsTool.ListFiles)

	writeTool, _ := utils.InferTool[WriteFileRequest, WriteFileResponse](
		"write_file",
		"Write content to a file",
		fsTool.WriteFile)

	readTool, _ := utils.InferTool[ReadFileRequest, ReadFileResponse](
		"read_file",
		"Read content from a file",
		fsTool.ReadFile)

	return []tool.BaseTool{listTool, writeTool, readTool}
}

func (t *FileSystemTool) ListFiles(ctx context.Context, req ListFilesRequest) (ListFilesResponse, error) {
	var files []string

	err := filepath.WalkDir(req.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only include files in the immediate directory (not subdirectories)
		if filepath.Dir(path) == req.Directory && !d.IsDir() {
			files = append(files, d.Name())
		}

		// Skip subdirectories
		if d.IsDir() && path != req.Directory {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return ListFilesResponse{
			Files: []string{},
			Error: fmt.Sprintf("Failed to list files: %v", err),
		}, nil
	}

	return ListFilesResponse{
		Files: files,
		Error: "",
	}, nil
}

func (t *FileSystemTool) WriteFile(ctx context.Context, req WriteFileRequest) (WriteFileResponse, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(req.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return WriteFileResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		}, nil
	}

	// Write file
	err := os.WriteFile(req.FilePath, []byte(req.Content), 0644)
	if err != nil {
		return WriteFileResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write file: %v", err),
		}, nil
	}

	return WriteFileResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (t *FileSystemTool) ReadFile(ctx context.Context, req ReadFileRequest) (ReadFileResponse, error) {
	content, err := os.ReadFile(req.FilePath)
	if err != nil {
		return ReadFileResponse{
			Content: "",
			Error:   fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}

	return ReadFileResponse{
		Content: string(content),
		Error:   "",
	}, nil
}
