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
	"strconv"
	"strings"

	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/grpc"
	"github.com/knstch/course/internal/app/grpc/grpcvideo"
	cdnerrors "github.com/knstch/course/internal/app/services/cdn_errors"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
	"golang.org/x/sync/errgroup"
)

var (
	ErrUnautharizedAccess = errors.New("доступ к курсу запрещен")
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
	ToggleHiddenStatus(ctx context.Context, courseId int) *courseError.CourseError
	DeleteModule(ctx context.Context, moduleId string) *courseError.CourseError
	DeleteLesson(ctx context.Context, lessonId string) *courseError.CourseError
	GetCourseByName(ctx context.Context, name string) (*dto.Course, *courseError.CourseError)
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

func (manager ContentManagementServcie) AddCourse(ctx context.Context,
	name, description, cost, discount string,
	formFileHeader *multipart.FileHeader, file *multipart.File) (*uint, *courseError.CourseError) {
	if err := validation.NewCourseToValidate(name, description, cost, discount, formFileHeader.Filename).Validate(ctx); err != nil {
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

	if err := writer.Close(); err != nil {
		return nil, courseError.CreateError(err, 500)
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/uploadCourseImage", manager.cdnHost), body)
	if err != nil {
		return nil, courseError.CreateError(err, 11040)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("ADMIN-API-KEY", manager.adminApiKey)

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

func (manager ContentManagementServcie) AddModule(ctx context.Context, module *entity.Module) (*uint, *courseError.CourseError) {
	if err := validation.NewModuleToValidate(module.Name, module.Description, module.CourseName, *module.Position).
		Validate(ctx); err != nil {
		return nil, err
	}

	id, err := manager.contentManager.CreateModule(ctx, module.Name, module.Description, module.CourseName, *module.Position)
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
	courseName string,
	preview *multipart.FileHeader,
	previewFile *multipart.File,
) (*uint, *courseError.CourseError) {
	if err := validation.NewLessonToValidate(
		name, description, moduleName, preview.Filename, video.Filename, position, courseName,
	).Validate(ctx); err != nil {
		return nil, err
	}

	if err := manager.contentManager.CheckIfLessonCanBeCreated(ctx, name, moduleName, position, courseName); err != nil {
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

func (manager ContentManagementServcie) GetCourseInfo(ctx context.Context, id, name, descr, cost, discount, page, limit string) (*entity.CourseInfoWithPagination, *courseError.CourseError) {
	var isCoursePurchased bool
	if id != "" {
		if ctx.Value("userId") == nil && ctx.Value("adminId") == nil {
			return nil, courseError.CreateError(ErrUnautharizedAccess, 13004)
		}

		if err := validation.NewStringIdToValidate(id).Validate(ctx); err != nil {
			return nil, err
		}

		courses, err := manager.contentManager.GetUserCourses(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range courses {
			if fmt.Sprint(v.CourseId) == id {
				isCoursePurchased = true
			}
		}
	}

	if ctx.Value("adminId") != nil {
		isCoursePurchased = true
	}

	if err := validation.NewCourseQueryToValidate(name, descr, cost, discount, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	offset := pageInt * limitInt

	courseInfo, err := manager.contentManager.GetCourse(ctx, id, name, descr, cost, discount, limitInt, offset, isCoursePurchased)
	if err != nil {
		return nil, err
	}

	return &entity.CourseInfoWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(courseInfo),
			PagesCount: len(courseInfo) / limitInt,
		},
		CourseInfo: courseInfo,
	}, nil
}

func (manager ContentManagementServcie) GetModulesInfo(ctx context.Context,
	name, description, courseName, page, limit string) (*entity.ModuleInfoWithPagination, *courseError.CourseError) {
	if err := validation.NewModuleQueryToValidate(name, description, courseName, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	offset := pageInt * limitInt

	var isPurchased bool
	if ctx.Value("userId") != nil && courseName != "" {
		userCourses, err := manager.contentManager.GetUserCourses(ctx)
		if err != nil {
			return nil, err
		}

		course, err := manager.contentManager.GetCourseByName(ctx, courseName)
		if err != nil {
			return nil, err
		}

		if course != nil {
			for _, v := range userCourses {
				if v.CourseId == course.ID {
					isPurchased = true
					break
				}
			}
		}
	}

	if ctx.Value("adminId") != nil {
		isPurchased = true
	}

	modules, err := manager.contentManager.GetModules(ctx, name, description, courseName, limitInt, offset, isPurchased)
	if err != nil {
		return nil, err
	}

	return &entity.ModuleInfoWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(modules),
			PagesCount: len(modules) / limitInt,
		},
		ModuleInfo: modules,
	}, nil
}

func (manager ContentManagementServcie) GetLessonsInfo(ctx context.Context,
	name, description, moduleName, courseName, page, limit string) (
	*entity.LessonsInfoWithPagination, *courseError.CourseError) {
	if err := validation.NewLessonsQueryToValidate(name, description, courseName, moduleName, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	offset := pageInt * limitInt

	var isPurchased bool
	if ctx.Value("userId") != nil && courseName != "" {
		userCourses, err := manager.contentManager.GetUserCourses(ctx)
		if err != nil {
			return nil, err
		}

		course, err := manager.contentManager.GetCourseByName(ctx, courseName)
		if err != nil {
			return nil, err
		}

		if course != nil {
			for _, v := range userCourses {
				if v.CourseId == course.ID {
					isPurchased = true
					break
				}
			}
		}
	}

	if ctx.Value("adminId") != nil {
		isPurchased = true
	}

	lessons, err := manager.contentManager.GetLessons(ctx, name, description, moduleName, courseName, limitInt, offset, isPurchased)
	if err != nil {
		return nil, err
	}

	return &entity.LessonsInfoWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(lessons),
			PagesCount: len(lessons) / limitInt,
		},
		LessonInfo: lessons,
	}, nil
}

func (manager ContentManagementServcie) ManageCourse(ctx context.Context,
	courseId, name, description, cost, discount string,
	formFileHeader *multipart.FileHeader,
	file *multipart.File, fileNotExists bool) *courseError.CourseError {
	if err := validation.NewEditCoruseToValidate(name, description, cost, discount, courseId).Validate(ctx); err != nil {
		return err
	}

	var (
		path         *string
		err          *courseError.CourseError
		uintCost     *uint
		uintDiscount *uint
	)

	if !fileNotExists {
		if err := validation.NewPreviewFileNameToValidate(formFileHeader.Filename).Validate(ctx); err != nil {
			return err
		}

		readyName := manager.prepareFileName(formFileHeader.Filename)

		path, err = manager.sendPhoto(file, readyName)
		if err != nil {
			return err
		}
	}

	if cost != "" {
		intCost, _ := strconv.Atoi(cost)
		bufferCost := uint(intCost)
		uintCost = &bufferCost
	}
	if discount != "" {
		intDiscount, _ := strconv.Atoi(discount)
		bufferDiscount := uint(intDiscount)
		uintCost = &bufferDiscount
	}

	if err := manager.contentManager.EditCourse(ctx, courseId, name, description, path, uintCost, uintDiscount); err != nil {
		return err
	}

	return nil
}

func (manager ContentManagementServcie) ManageModule(ctx context.Context, module *entity.Module) *courseError.CourseError {
	if err := validation.NewEditModuleToValidate(module.Name, module.Description, *module.Position, module.ModuleId).
		Validate(ctx); err != nil {
		return err
	}

	if err := manager.contentManager.EditModule(ctx, module.Name, module.Description, module.Position, module.ModuleId); err != nil {
		return err
	}

	return nil
}

func (manager ContentManagementServcie) ManageLesson(ctx context.Context,
	video *multipart.FileHeader,
	name string,
	description string,
	position string,
	lessonId string,
	preview *multipart.FileHeader,
	previewFile *multipart.File,
	videoNotExists bool,
	previewNotExists bool,
) *courseError.CourseError {
	if err := validation.NewEditLessonToValidate(name, description, position, lessonId).Validate(ctx); err != nil {
		return err
	}

	var (
		previewPath *string
		videoPath   *string
		err         *courseError.CourseError
	)

	if !previewNotExists {
		if err := validation.NewPreviewFileNameToValidate(preview.Filename).Validate(ctx); err != nil {
			return err
		}

		readyName := manager.prepareFileName(preview.Filename)

		previewPath, err = manager.sendPhoto(previewFile, readyName)
		if err != nil {
			return err
		}
	}

	if !videoNotExists {
		if err := validation.NewVideoFileNameToValidate(video.Filename).Validate(ctx); err != nil {
			return err
		}

		videoPath, err = manager.sendVideo(ctx, video)
		if err != nil {
			return err
		}
	}

	if err := manager.contentManager.EditLesson(ctx, name, description, position, lessonId, videoPath, previewPath); err != nil {
		return err
	}

	return nil
}

func (manager ContentManagementServcie) ManageShowStatus(ctx context.Context, courseId int) *courseError.CourseError {
	if err := validation.NewIdToValidate(courseId).Validate(ctx); err != nil {
		return err
	}

	if err := manager.contentManager.ToggleHiddenStatus(ctx, courseId); err != nil {
		return err
	}

	return nil
}

func (manager ContentManagementServcie) RemoveModule(ctx context.Context, moduleId string) *courseError.CourseError {
	if err := validation.NewStringIdToValidate(moduleId).Validate(ctx); err != nil {
		return err
	}

	if err := manager.contentManager.DeleteModule(ctx, moduleId); err != nil {
		return err
	}

	return nil
}

func (manager ContentManagementServcie) RemoveLesson(ctx context.Context, lessonId string) *courseError.CourseError {
	if err := validation.NewStringIdToValidate(lessonId).Validate(ctx); err != nil {
		return err
	}

	if err := manager.contentManager.DeleteLesson(ctx, lessonId); err != nil {
		return err
	}

	return nil
}
