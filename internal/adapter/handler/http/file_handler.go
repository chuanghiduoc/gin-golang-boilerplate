package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend-gin/internal/adapter/handler/http/middleware"
	"backend-gin/internal/adapter/handler/http/response"
	"backend-gin/internal/domain/entity"
	"backend-gin/internal/usecase/file"
)

type FileHandler struct {
	fileService file.Service
}

func NewFileHandler(fileService file.Service) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// Upload godoc
// @Summary Upload a file
// @Description Upload a file to storage
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Param path formData string false "Storage path"
// @Success 201 {object} response.Response{data=file.FileResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/files [post]
func (h *FileHandler) Upload(c *gin.Context) {
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

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	path := c.PostForm("path")

	result, err := h.fileService.Upload(c.Request.Context(), id, fileHeader, path)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Created(c, result)
}

// GetFile godoc
// @Summary Get file by ID
// @Description Get a file's metadata by its ID
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 200 {object} response.Response{data=file.FileResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid file id")
		return
	}

	result, err := h.fileService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Success(c, result)
}

// ListMyFiles godoc
// @Summary List my files
// @Description Get a paginated list of current user's files
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} response.PaginatedResponse{data=[]file.FileResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/files/me [get]
func (h *FileHandler) ListMyFiles(c *gin.Context) {
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

	var req file.ListFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.fileService.ListByUserID(c.Request.Context(), id, &req)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.PaginatedWithMeta(c, result.Files, result.Meta)
}

// ListAllFiles godoc
// @Summary List all files (Admin only)
// @Description Get a paginated list of all files
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} response.PaginatedResponse{data=[]file.FileResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/files [get]
func (h *FileHandler) ListAllFiles(c *gin.Context) {
	var req file.ListFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.fileService.List(c.Request.Context(), &req)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.PaginatedWithMeta(c, result.Files, result.Meta)
}

// DeleteFile godoc
// @Summary Delete a file
// @Description Delete a file by ID (users can delete their own files, admins can delete any)
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid file id")
		return
	}

	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		response.Unauthorized(c, "user not found in context")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		response.InternalError(c)
		return
	}

	isAdmin := middleware.GetUserRole(c) == entity.RoleAdmin

	if err := h.fileService.Delete(c.Request.Context(), id, uid, isAdmin); err != nil {
		response.Err(c, err)
		return
	}

	response.NoContent(c)
}
