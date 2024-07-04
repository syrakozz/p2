package llm

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"disruptive/lib/common"
	"disruptive/lib/openai"
)

// ListFiles lists all files uploaded to OpenAI.
func ListFiles(ctx context.Context) error {
	logCtx := slog.With()
	files, err := openai.ListFiles(ctx, logCtx)
	if err != nil {
		return err
	}

	common.P(files)
	return nil
}

// UploadFile uploads a file to OpenAI.
func UploadFile(ctx context.Context, filePath, purpose string) error {
	logCtx := slog.With("file_path", filePath, "purpose", purpose)
	logCtx.Info("Uploading file...")

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	res, err := openai.UploadFile(ctx, logCtx, filepath.Base(f.Name()), f, "fine-tune")
	if err != nil {
		return err
	}

	logCtx.Info("file uploaded", "id", res.ID)
	return nil
}

// DeleteFile delete a file uploaded to OpenAI
func DeleteFile(ctx context.Context, fileID string) error {
	logCtx := slog.With("file_id", fileID)
	return openai.DeleteFile(ctx, logCtx, fileID)
}
