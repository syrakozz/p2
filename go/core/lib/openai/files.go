package openai

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"disruptive/lib/common"
)

type _createFileResponse struct {
	Object        string  `json:"object"`
	ID            string  `json:"id"`
	Purpose       string  `json:"purpose"`
	Filename      string  `json:"filename"`
	Bytes         int     `json:"bytes"`
	CreatedAt     int64   `json:"created_at"`
	Status        string  `json:"status"`
	StatusDetails *string `json:"status_details"`
}

// CreateFileResponse contains the CreateFile API response structure.
type CreateFileResponse struct {
	ID            string    `json:"id"`
	Purpose       string    `json:"purpose"`
	Filename      string    `json:"filename"`
	Bytes         int       `json:"bytes"`
	CreatedAt     time.Time `json:"created_at"`
	Status        string    `json:"status"`
	StatusDetails *string   `json:"status_details,omitempty"`
}

type _filesResponse struct {
	Data []struct {
		ID        string `json:"id"`
		Object    string `json:"object"`
		Bytes     int    `json:"bytes"`
		CreatedAt int64  `json:"created_at"`
		Filename  string `json:"filename"`
		Purpose   string `json:"purpose"`
	} `json:"data"`
	Object string `json:"object"`
}

// FileResponse contains the File API response stucture.
type FileResponse struct {
	ID        string    `json:"id"`
	Bytes     int       `json:"bytes"`
	CreatedAt time.Time `json:"created_at"`
	Filename  string    `json:"filename"`
	Purpose   string    `json:"purpose"`
}

func renderToFileResponses(_res *_filesResponse) []FileResponse {
	res := make([]FileResponse, len(_res.Data))

	for i, d := range _res.Data {
		res[i] = FileResponse{
			ID:        d.ID,
			Bytes:     d.Bytes,
			CreatedAt: time.Unix(d.CreatedAt, 0),
			Filename:  d.Filename,
			Purpose:   d.Purpose,
		}
	}

	return res
}

func renderToCreateFileResponse(_res *_createFileResponse) CreateFileResponse {
	return CreateFileResponse{
		ID:            _res.ID,
		Purpose:       _res.Purpose,
		Filename:      _res.Filename,
		Bytes:         _res.Bytes,
		CreatedAt:     time.Unix(_res.CreatedAt, 0),
		Status:        _res.Status,
		StatusDetails: _res.StatusDetails,
	}
}

// ListFiles returns a list of OpenAI files.
func ListFiles(ctx context.Context, logCtx *slog.Logger) ([]FileResponse, error) {
	fid := slog.String("fid", "openai.ListFiles")

	res, err := Resty.R().
		SetContext(ctx).
		SetResult(&_filesResponse{}).
		Get(filesEndpoint)

	if err != nil {
		logCtx.Error("openai files endpoint failed", fid, "error", err)
		return nil, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return nil, common.ErrUnauthorized
	}

	return renderToFileResponses(res.Result().(*_filesResponse)), nil
}

// UploadFile uploads a file to OpenAI.
func UploadFile(ctx context.Context, logCtx *slog.Logger, name string, reader io.Reader, purpose string) (CreateFileResponse, error) {
	fid := slog.String("fid", "openai.UploadFinetunes")

	res, err := Resty.R().
		SetContext(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetFileReader("file", name, reader).
		SetFormData(map[string]string{"purpose": purpose}).
		SetResult(&_createFileResponse{}).
		Post(filesEndpoint)

	if err != nil {
		logCtx.Error("openai files endpoint failed", fid, "error", err)
		return CreateFileResponse{}, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return CreateFileResponse{}, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai files endpoint failed", fid, "status", res.Status(), "message", string(res.Body()))
		return CreateFileResponse{}, errors.New(res.Status())
	}

	return renderToCreateFileResponse(res.Result().(*_createFileResponse)), nil
}

// DeleteFile deletes an OpenAI file.
func DeleteFile(ctx context.Context, logCtx *slog.Logger, fileID string) error {
	fid := slog.String("fid", "openai.DeleteFile")

	res, err := Resty.R().
		SetContext(ctx).
		SetPathParam("file_id", fileID).
		Delete(fileEndpoint)

	if err != nil {
		logCtx.Error("openai file endpoint failed", fid, "error", err)
		return err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai file endpoint failed", fid, "status", res.Status(), "message", errorMessage(res.Body()))
		return errors.New(res.Status())
	}

	return nil
}
