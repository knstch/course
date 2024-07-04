package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

func (h Handlers) FindUsersByFilters(ctx *gin.Context) {
	firstName := ctx.Query("firstName")
	surname := ctx.Query("surname")
	phoneNumber := ctx.Query("phoneNumber")
	email := ctx.Query("email")
	active := ctx.Query("active")
	verified := ctx.Query("verified")
	courseName := ctx.Query("courseName")
	banned := ctx.Query("banned")
	page := ctx.Query("page")
	limit := ctx.Query("limit")
	users, err := h.userManagementService.RetreiveUsersByFilters(ctx, firstName, surname,
		phoneNumber, email, active, verified, courseName, banned, page, limit)
	if err != nil {
		h.logger.Error(
			fmt.Sprintf("ошибка при поиске по фильтрам: firstName - %v, surname - %v, phoneNumber - %v, email - %v, active - %v, verified - %v, courseName - %v, banned - %v, page - %v, limit- %v", firstName, surname,
				phoneNumber, email, active, verified, courseName, banned, page, limit), "FindUsersByFilters", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователи успешно получены по фильртрам админом с ID: %d",
		ctx.Value("adminId").(uint)),
		"FindUsersByFilters",
		fmt.Sprintf("фильтры: firstName - %v, surname - %v, phoneNumber - %v, email - %v, active - %v, verified - %v, courseName - %v, banned - %v, page - %v, limit- %v",
			firstName, surname, phoneNumber, email, active, verified, courseName, banned, page, limit))

	ctx.JSON(http.StatusOK, users)
}

func (h Handlers) BanUser(ctx *gin.Context) {
	Id := entity.NewId(nil)
	if err := ctx.ShouldBindJSON(&Id); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "BanUser", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.userManagementService.DeactivateUser(ctx, Id.Id); err != nil {
		h.logger.Error("ошибка при блокировке пользователя", "BanUser", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователь успешно забанен админом с ID: %v",
		ctx.Value("adminId").(uint)), "BanUser", fmt.Sprintf("userId: %d", Id.Id))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пользователь успешно заблокирован"))
}

func (h Handlers) UnbanUser(ctx *gin.Context) {
	Id := entity.NewId(nil)
	if err := ctx.ShouldBindJSON(&Id); err != nil {
		h.logger.Error("не получилось обработать тело запроса", "UnbanUser", err.Error(), 10101)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, courseError.CreateError(errBrokenJSON, 10101))
		return
	}

	if err := h.userManagementService.ActivateUser(ctx, Id.Id); err != nil {
		h.logger.Error("ошибка при разблокировке пользователя", "UnbanUser", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователь успешно разблокирован админом с ID: %v",
		ctx.Value("adminId").(uint)), "UnbanUser", fmt.Sprintf("userId: %d", Id.Id))

	ctx.JSON(http.StatusOK, entity.CreateSuccessResponse("пользователь успешно разблокирован"))
}

func (h Handlers) GetUserById(ctx *gin.Context) {
	id := ctx.Query("id")
	user, err := h.userManagementService.RetreiveUserById(ctx, id)
	if err != nil {
		h.logger.Error(fmt.Sprintf("ошибка при получении пользователя по ID: %v", id), "GetUserById", err.Message, err.Code)
		if err.Code == 400 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
		if err.Code == 11101 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	h.logger.Info(fmt.Sprintf("пользователь успешно получен админом с ID: %d", ctx.Value("adminId").(uint)), "GetUserById", fmt.Sprintf("userId: %v", id))

	ctx.JSON(http.StatusOK, user)
}
