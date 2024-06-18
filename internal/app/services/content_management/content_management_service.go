package contentmanagement

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/grpc"
	"github.com/knstch/course/internal/app/grpc/grpcvideo"
	cdnerrors "github.com/knstch/course/internal/app/services/cdn_errors"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
	"golang.org/x/sync/errgroup"
)

type ContentManagementServcie struct {
	contentManager ContentManager
	adminApiKey    string
	cdnHost        string
	client         *http.Client
	grpcClient     *grpc.GrpcClient
}

type ContentManager interface {
	CreateCourse(ctx context.Context, name, description, cost, discount, path string) (*uint, *courseError.CourseError)
	CreateModule(ctx context.Context, name, description string, position, courseId uint) (*uint, *courseError.CourseError)
	CheckIfLessonCanBeCreated(ctx context.Context, name, moduleName, position string) *courseError.CourseError
	CreateLesson(ctx context.Context, name, moduleName, description, position, videoPath, previewPath string) (*uint, *courseError.CourseError)
}

func NewContentManagementServcie(manager ContentManager, config *config.Config, client *http.Client, grpcClient *grpc.GrpcClient) ContentManagementServcie {
	return ContentManagementServcie{
		contentManager: manager,
		adminApiKey:    config.CdnAdminApiKey,
		cdnHost:        config.CdnHost,
		client:         client,
		grpcClient:     grpcClient,
	}
}

func (manager ContentManagementServcie) AddCourse(ctx context.Context, name, description, cost, discount string, formFileHeader *multipart.FileHeader, file *multipart.File) (*uint, *courseError.CourseError) {
	if err := validation.NewCourseToValidate(name, description, formFileHeader.Filename, cost, discount).Validate(ctx); err != nil {
		return nil, err
	}

	readyName := manager.prepareFileName(formFileHeader.Filename)

	path, err := manager.sendPhoto(file, readyName)
	if err != nil {
		return nil, err
	}

	id, err := manager.contentManager.CreateCourse(ctx, name, description, cost, discount, *path)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (manager ContentManagementServcie) prepareFileName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
}

func (manager ContentManagementServcie) sendPhoto(file *multipart.File, fileName string) (*string, *courseError.CourseError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile("photo", fileName)
	if err != nil {
		return nil, courseError.CreateError(err, 11031)
	}

	_, err = io.Copy(formFile, *file)
	if err != nil {
		return nil, courseError.CreateError(err, 11042)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/course", manager.cdnHost), body)
	if err != nil {
		return nil, courseError.CreateError(err, 11040)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("ADMIN-API-KEY", manager.adminApiKey)

	resp, err := manager.client.Do(req)
	if err != nil {
		return nil, courseError.CreateError(cdnerrors.ErrCdnNotResponding, 11041)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, courseError.CreateError(err, 11042)
	}

	cdnResponse := entity.NewCdnResponse()
	if err := json.Unmarshal(respBody, &cdnResponse); err != nil {
		return nil, courseError.CreateError(err, 10101)
	}

	if cdnResponse.Err != nil {
		if cdnResponse.Code == 403 {
			return nil, courseError.CreateError(cdnerrors.ErrFailedAuth, 11050)
		}
		if cdnResponse.Code == 400 {
			return nil, courseError.CreateError(cdnerrors.ErrBadFile, 11105)
		}
		if cdnResponse.Code == 1000 {
			return nil, courseError.CreateError(cdnerrors.ErrCdnFailture, 11051)
		}
	}

	return &cdnResponse.Path, nil
}

func (manager ContentManagementServcie) AddModule(ctx context.Context, module *entity.Module) (*uint, *courseError.CourseError) {
	if err := validation.NewModuleToValidate(module.Name, module.Description, module.Position, module.CourseId).
		Validate(ctx); err != nil {
		return nil, err
	}

	id, err := manager.contentManager.CreateModule(ctx, module.Name, module.Description, module.Position, module.CourseId)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (manager ContentManagementServcie) AddLesson(
	ctx context.Context,
	video *multipart.FileHeader,
	name string,
	moduleName string,
	description string,
	position string,
	preview *multipart.FileHeader,
	previewFile *multipart.File,
) (*uint, *courseError.CourseError) {
	if err := validation.NewLessonToValidate(
		name, description, moduleName, preview.Filename, video.Filename, position,
	).Validate(ctx); err != nil {
		return nil, err
	}

	if err := manager.contentManager.CheckIfLessonCanBeCreated(ctx, name, moduleName, position); err != nil {
		return nil, err
	}

	var (
		videoPath    *string
		sendVideoErr *courseError.CourseError

		previewPath  *string
		sendPhotoErr *courseError.CourseError
	)

	g, errGroupCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		videoPath, sendVideoErr = manager.sendVideo(errGroupCtx, video)
		if sendVideoErr != nil {
			return sendVideoErr.Error
		}
		return nil
	})

	g.Go(func() error {
		readyPreviewFileName := manager.prepareFileName(preview.Filename)
		previewPath, sendPhotoErr = manager.sendPhoto(previewFile, readyPreviewFileName)
		if sendPhotoErr != nil {
			return sendPhotoErr.Error
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		if sendPhotoErr != nil {
			return nil, sendPhotoErr
		}
		if sendVideoErr != nil {
			return nil, sendVideoErr
		}
	}

	lessonId, err := manager.contentManager.CreateLesson(ctx, name, moduleName, description, position, *videoPath, *previewPath)
	if err != nil {
		return nil, err
	}

	return lessonId, nil
}

func (manager ContentManagementServcie) sendVideo(ctx context.Context, file *multipart.FileHeader) (*string, *courseError.CourseError) {
	readyName := manager.prepareFileName(file.Filename)

	stream, err := manager.grpcClient.Client.UploadVideo(ctx)
	if err != nil {
		return nil, courseError.CreateError(err, 14002)
	}

	fileSize := file.Size

	buffer := make([]byte, fileSize)

	fileReader, err := file.Open()
	if err != nil {
		return nil, courseError.CreateError(err, 11042)
	}

	for {
		bytesRead, err := fileReader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, courseError.CreateError(err, 11042)
		}

		if err := stream.Send(&grpcvideo.UploadVideoRequest{
			Content: buffer[:bytesRead],
			Name:    readyName,
		}); err != nil {
			return nil, courseError.CreateError(err, 14002)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return nil, courseError.CreateError(err, 14002)
	}

	return &res.Path, nil
}
