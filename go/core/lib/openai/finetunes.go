package openai

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"disruptive/lib/common"
)

// FinetuneRequest contains the Finetune API request structure.
type FinetuneRequest struct {
	TrainingFile                 string    `json:"training_file"`
	ValidationFile               string    `json:"validation_file,omitempty"`
	Model                        string    `json:"model,omitempty"`
	NEpochs                      int       `json:"n_epochs,omitempty"`
	BatchSize                    int       `json:"batch_size,omitempty"`
	LearningRateMultiplier       float64   `json:"learning_rate_multiplier,omitempty"`
	PromptLossWeight             float64   `json:"prompt_loss_weight,omitempty"`
	ComputeClassificationMetrics bool      `json:"compute_classification_metrics,omitempty"`
	ClassificationNClasses       int       `json:"classification_n_classes,omitempty"`
	ClassificationPositiveClass  string    `json:"classification_positive_class,omitempty"`
	ClassificationBetas          []float64 `json:"classification_betas,omitempty"`
	Suffix                       string    `json:"suffix,omitempty"`
}

type _finetuneEvent struct {
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// FinetuneEvent contains the Finetune event structure.
type FinetuneEvent struct {
	CreatedAt time.Time `json:"created_at"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type _finetuneFile struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type _finetuneHyperparams struct {
	BatchSize              int     `json:"batch_size"`
	LearningRateMultiplier float64 `json:"learning_rate_multiplier"`
	NEpochs                int     `json:"n_epochs"`
	PromptLossWeight       float64 `json:"prompt_loss_weight"`
}

type _finetuneData struct {
	ID              string               `json:"id"`
	Object          string               `json:"object"`
	Model           string               `json:"model"`
	CreatedAt       int64                `json:"created_at"`
	Events          []_finetuneEvent     `json:"events"`
	FineTunedModel  *string              `json:"fine_tuned_model"`
	Hyperparams     _finetuneHyperparams `json:"hyperparams"`
	OrganizationID  string               `json:"organization_id"`
	ResultFiles     []_finetuneFile      `json:"result_files"`
	Status          string               `json:"status"`
	ValidationFiles []_finetuneFile      `json:"validation_files"`
	TrainingFiles   []_finetuneFile      `json:"training_files"`
	UpdatedAt       int64                `json:"updated_at"`
}

type _finetuneResponse struct {
	Object string          `json:"object"`
	Data   []_finetuneData `json:"data"`
}

// FinetuneFile contains the Finetune file structure.
type FinetuneFile struct {
	ID        string    `json:"id"`
	Bytes     int       `json:"bytes"`
	CreatedAt time.Time `json:"created_at"`
	Filename  string    `json:"filename"`
	Purpose   string    `json:"purpose"`
}

// FinetuneResponse contains the Finetune API request structure.
type FinetuneResponse struct {
	ID              string               `json:"id"`
	Model           string               `json:"model"`
	CreatedAt       time.Time            `json:"created_at"`
	Events          []FinetuneEvent      `json:"events,omitempty"`
	FineTunedModel  *string              `json:"fine_tuned_model,omitempty"`
	Hyperparams     _finetuneHyperparams `json:"hyperparams"`
	OrganizationID  string               `json:"organization_id,omitempty"`
	ResultFiles     []FinetuneFile       `json:"result_files"`
	Status          string               `json:"status"`
	ValidationFiles []FinetuneFile       `json:"validation_files,omitempty"`
	TrainingFiles   []FinetuneFile       `json:"training_files,omitempty"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

func renderToFinetuneResponse(_res *_finetuneData) FinetuneResponse {
	events := make([]FinetuneEvent, len(_res.Events))
	for i, e := range _res.Events {
		events[i] = FinetuneEvent{
			CreatedAt: time.Unix(e.CreatedAt, 0),
			Level:     e.Level,
			Message:   e.Message,
		}
	}

	rf := make([]FinetuneFile, len(_res.ResultFiles))
	for i, f := range _res.ResultFiles {
		rf[i] = FinetuneFile{
			ID:        f.ID,
			Bytes:     f.Bytes,
			CreatedAt: time.Unix(f.CreatedAt, 0),
			Filename:  f.Filename,
			Purpose:   f.Purpose,
		}
	}

	vf := make([]FinetuneFile, len(_res.ValidationFiles))
	for i, f := range _res.ValidationFiles {
		vf[i] = FinetuneFile{
			ID:        f.ID,
			Bytes:     f.Bytes,
			CreatedAt: time.Unix(f.CreatedAt, 0),
			Filename:  f.Filename,
			Purpose:   f.Purpose,
		}
	}

	tf := make([]FinetuneFile, len(_res.TrainingFiles))
	for i, f := range _res.TrainingFiles {
		tf[i] = FinetuneFile{
			ID:        f.ID,
			Bytes:     f.Bytes,
			CreatedAt: time.Unix(f.CreatedAt, 0),
			Filename:  f.Filename,
			Purpose:   f.Purpose,
		}
	}

	res := FinetuneResponse{
		ID:              _res.ID,
		Model:           _res.Model,
		CreatedAt:       time.Unix(_res.CreatedAt, 0),
		Events:          events,
		FineTunedModel:  _res.FineTunedModel,
		Hyperparams:     _res.Hyperparams,
		OrganizationID:  _res.OrganizationID,
		ResultFiles:     rf,
		Status:          _res.Status,
		ValidationFiles: vf,
		TrainingFiles:   tf,
		UpdatedAt:       time.Unix(_res.UpdatedAt, 0),
	}

	return res
}

func renderToFinetuneResponses(_res *_finetuneResponse) []FinetuneResponse {
	res := make([]FinetuneResponse, len(_res.Data))

	for i, d := range _res.Data {
		rf := make([]FinetuneFile, len(d.ResultFiles))
		for i, f := range d.ResultFiles {
			rf[i] = FinetuneFile{
				ID:        f.ID,
				Bytes:     f.Bytes,
				CreatedAt: time.Unix(f.CreatedAt, 0),
				Filename:  f.Filename,
				Purpose:   f.Purpose,
			}
		}

		vf := make([]FinetuneFile, len(d.ValidationFiles))
		for i, f := range d.ValidationFiles {
			vf[i] = FinetuneFile{
				ID:        f.ID,
				Bytes:     f.Bytes,
				CreatedAt: time.Unix(f.CreatedAt, 0),
				Filename:  f.Filename,
				Purpose:   f.Purpose,
			}
		}

		tf := make([]FinetuneFile, len(d.TrainingFiles))
		for i, f := range d.TrainingFiles {
			tf[i] = FinetuneFile{
				ID:        f.ID,
				Bytes:     f.Bytes,
				CreatedAt: time.Unix(f.CreatedAt, 0),
				Filename:  f.Filename,
				Purpose:   f.Purpose,
			}
		}

		res[i] = FinetuneResponse{
			ID:              d.ID,
			Model:           d.Model,
			CreatedAt:       time.Unix(d.CreatedAt, 0),
			FineTunedModel:  d.FineTunedModel,
			Hyperparams:     d.Hyperparams,
			OrganizationID:  d.OrganizationID,
			ResultFiles:     rf,
			Status:          d.Status,
			ValidationFiles: vf,
			TrainingFiles:   tf,
			UpdatedAt:       time.Unix(d.UpdatedAt, 0),
		}
	}

	return res
}

// GetFinetunes returns basic information for all Finetune models.
func GetFinetunes(ctx context.Context, logCtx *slog.Logger) ([]FinetuneResponse, error) {
	fid := slog.String("fid", "openai.GetFinetunes")

	res, err := Resty.R().
		SetContext(ctx).
		SetResult(&_finetuneResponse{}).
		Get(finetunesEndpoint)

	if err != nil {
		logCtx.Error("openai fine-tunes endpoint failed", fid, "error", err)
		return nil, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return nil, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai fine-tunes endpoint failed", fid, "status", res.Status())
		return nil, errors.New(res.Status())
	}

	return renderToFinetuneResponses(res.Result().(*_finetuneResponse)), nil
}

// GetFinetune returns detail information about a finetune model.
func GetFinetune(ctx context.Context, logCtx *slog.Logger, finetuneID string) (FinetuneResponse, error) {
	fid := slog.String("fid", "openai.GetFinetune")

	res, err := Resty.R().
		SetContext(ctx).
		SetPathParam("finetune_id", finetuneID).
		SetResult(&_finetuneData{}).
		Get(finetuneEndpoint)

	if err != nil {
		logCtx.Error("openai fine-tune endpoint failed", fid, "error", err)
		return FinetuneResponse{}, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return FinetuneResponse{}, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai fine-tune endpoint failed", fid, "status", res.Status())
		return FinetuneResponse{}, errors.New(res.Status())
	}

	return renderToFinetuneResponse(res.Result().(*_finetuneData)), nil
}

// PostFinetune creates a new Finetune model.
func PostFinetune(ctx context.Context, logCtx *slog.Logger, request FinetuneRequest) (FinetuneResponse, error) {
	fid := slog.String("fid", "openai.PostFinetune")

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&_finetuneData{}).
		Post(finetunesEndpoint)

	if err != nil {
		logCtx.Error("openai fine-tunes endpoint failed", fid, "error", err)
		return FinetuneResponse{}, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return FinetuneResponse{}, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai fine-tunes endpoint failed", fid, "status", res.Status())
		return FinetuneResponse{}, errors.New(res.Status())
	}

	return renderToFinetuneResponse(res.Result().(*_finetuneData)), nil
}
