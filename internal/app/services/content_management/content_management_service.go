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
)

type ContentManagementServcie struct {
	manager     ContentManager
	adminApiKey string
	cdnHost     string
	client      *http.Client
	grpcClient  *grpc.GrpcClient
}

type ContentManager interface {
	CreateCourse(ctx context.Context, name, description, cost, discount, path string) (*uint, *courseError.CourseError)
	CreateModule(ctx context.Context, name, description string, position, courseId uint) (*uint, *courseError.CourseError)
}

func NewContentManagementServcie(manager ContentManager, config *config.Config, client *http.Client, grpcClient *grpc.GrpcClient) ContentManagementServcie {
	return ContentManagementServcie{
		manager:     manager,
		adminApiKey: config.CdnAdminApiKey,
		cdnHost:     config.CdnHost,
		client:      client,
		grpcClient:  grpcClient,
	}
}

// func (manager ContentManagementServcie) AddLesson(ctx *gin.Context) {
// 	name := ctx.PostForm("name")
// 	description := ctx.PostForm("description")
// 	position := ctx.PostForm("position")
// }

func (manager ContentManagementServcie) AddCourse(ctx context.Context, name, description, cost, discount string, formFileHeader *multipart.FileHeader, file *multipart.File) (*uint, *courseError.CourseError) {
	if err := validation.NewCourseToValidate(name, description, formFileHeader.Filename, cost, discount).Validate(ctx); err != nil {
		return nil, err
	}

	readyName := manager.prepareFileName(formFileHeader.Filename)

	path, err := manager.sendCourse(file, readyName)
	if err != nil {
		return nil, err
	}

	id, err := manager.manager.CreateCourse(ctx, name, description, cost, discount, *path)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (manager ContentManagementServcie) prepareFileName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
}

func (manager ContentManagementServcie) sendCourse(file *multipart.File, fileName string) (*string, *courseError.CourseError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile("preview", fileName)
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

	id, err := manager.manager.CreateModule(ctx, module.Name, module.Description, module.Position, module.CourseId)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (manager ContentManagementServcie) AddLesson(ctx context.Context, file *multipart.File, header *multipart.FileHeader) (*uint, *courseError.CourseError) {
	rawData, err := io.ReadAll(*file)
	if err != nil {
		return nil, courseError.CreateError(err, 11042)
	}

	uploadVideoReq := grpcvideo.UploadVideoRequest{
		Content: rawData,
		Name:    header.Filename,
	}

	status, err := manager.grpcClient.Client.UploadVideo(ctx, &uploadVideoReq)
	if err != nil {
		fmt.Println("ERRRR: ", err.Error())
		return nil, courseError.CreateError(err, 14002)
	}

	fmt.Println("STATUS: ", status.Path, " SUC: ", status.Success)

	return nil, nil
}
