// contentmanagement содержит методы для менеджмента контента на платформе.
package contentmanagement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/grpc"
	"github.com/knstch/course/internal/app/grpc/grpcvideo"
	cdnerrors "github.com/knstch/course/internal/app/services/cdn_errors"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	ErrUnautharizedAccess = errors.New("доступ к курсу запрещен")
)

// ContentManagementServcie содержит данные для работы с CDN и получение контента.
type ContentManagementServcie struct {
	contentManager ContentManager
	adminApiKey    string
	cdnHost        string
	client         *http.Client
	grpcClient     *grpc.GrpcClient
}

type ContentManager interface {
	CreateCourse(ctx context.Context, name, description, cost, discount, path string) (*uint, *courseError.CourseError)
	CreateModule(ctx context.Context, name, description, courseName string, position uint) (*uint, *courseError.CourseError)
	CheckIfLessonCanBeCreated(ctx context.Context, name, moduleName, position, courseName string) *courseError.CourseError
	CreateLesson(ctx context.Context, name, moduleName, description, position, videoPath, previewPath string) (*uint, *courseError.CourseError)
	GetCourse(ctx context.Context, id, name, descr, cost, discount string, limit, offset int, isPurchased bool) ([]entity.CourseInfo, *courseError.CourseError)
	GetUserCourses(ctx context.Context) ([]dto.Order, *courseError.CourseError)
	GetModules(ctx context.Context, name, description, courseName string, limit, offset int, isPurchased bool) ([]entity.ModuleInfo, *courseError.CourseError)
	GetLessons(ctx context.Context, name, description, moduleName, courseName string, limit, offset int, isPurchased bool) ([]entity.LessonInfo, *courseError.CourseError)
	EditCourse(ctx context.Context, courseId, name, description string, previewUrl *string, cost, discount *uint) *courseError.CourseError
	EditModule(ctx context.Context, name, description string, position *uint, moduleId uint) *courseError.CourseError
	EditLesson(ctx context.Context, name, description, position, lessonId string, videoPath, previewPath *string) *courseError.CourseError
	ToggleHiddenStatus(ctx context.Context, courseId string) *courseError.CourseError
	DeleteModule(ctx context.Context, moduleId string) *courseError.CourseError
	DeleteLesson(ctx context.Context, lessonId string) *courseError.CourseError
	GetCourseByName(ctx context.Context, name string) (*dto.Course, *courseError.CourseError)
}

// NewContentManagementServcie - это билдер для сервиса контента.
func NewContentManagementServcie(manager ContentManager, config *config.Config, client *http.Client, grpcClient *grpc.GrpcClient) ContentManagementServcie {
	return ContentManagementServcie{
		contentManager: manager,
		adminApiKey:    config.CdnAdminApiKey,
		cdnHost:        config.CdnHost,
		client:         client,
		grpcClient:     grpcClient,
	}
}

// prepareFileName используется для подготовки названия файла к отправке, удаляя пробелы и заменяя их на _.
func (manager ContentManagementServcie) prepareFileName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
}

// sendPhoto отправляет превью на CDN и обрабатывает ошибку в случае неудачи.
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

	if err := writer.Close(); err != nil {
		return nil, courseError.CreateError(err, 500)
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/uploadCourseImage", manager.cdnHost), body)
	if err != nil {
		return nil, courseError.CreateError(err, 11040)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Admin-Api-Key", manager.adminApiKey)

	resp, err := manager.client.Do(req)
	if err != nil {
		return nil, courseError.CreateError(cdnerrors.ErrCdnNotResponding, 11041)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("ошибка при закрытии тела запроса: %v", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, courseError.CreateError(err, 11042)
	}

	cdnResponse := entity.NewCdnResponse()
	if err := json.Unmarshal(respBody, &cdnResponse); err != nil {
		return nil, courseError.CreateError(err, 10101)
	}

	if cdnResponse.Err != nil {
		switch {
		case cdnResponse.Code == 403:
			return nil, courseError.CreateError(cdnerrors.ErrFailedAuth, 11050)
		case cdnResponse.Code == 400:
			return nil, courseError.CreateError(cdnerrors.ErrBadFile, 11105)
		case cdnResponse.Code == 1000:
			return nil, courseError.CreateError(cdnerrors.ErrCdnFailture, 11051)
		default:
			return nil, courseError.CreateError(fmt.Errorf(*cdnResponse.Err), cdnResponse.Code)
		}
	}

	return &cdnResponse.Path, nil
}

// sendVideo используется для отправки видео по gRPC. Возвращает путь к контенту на CDN или ошибку.
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
