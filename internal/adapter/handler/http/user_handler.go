package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend-gin/internal/adapter/handler/http/middleware"
	"backend-gin/internal/adapter/handler/http/response"
	"backend-gin/internal/usecase/user"
)

type UserHandler struct {
	userService user.Service
}

func NewUserHandler(userService user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetMe godoc
// @Summary Get current user
// @Description Get the currently authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=user.UserResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		response.Unauthorized(c, "user not found in context")
		return
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalError(c)
		return
	}

	result, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Success(c, result)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=user.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	result, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Success(c, result)
}

// ListUsers godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} response.PaginatedResponse{data=[]user.UserResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req user.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.List(c.Request.Context(), &req)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.PaginatedWithMeta(c, result.Users, result.Meta)
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body user.CreateUserRequest true "User creation request"
// @Success 201 {object} response.Response{data=user.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Created(c, result)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update an existing user's information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body user.UpdateUserRequest true "User update request"
// @Success 200 {object} response.Response{data=user.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		response.Err(c, err)
		return
	}

	response.NoContent(c)
}
