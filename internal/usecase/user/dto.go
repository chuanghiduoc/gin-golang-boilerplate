package user

import (
	"time"

	"github.com/google/uuid"

	"backend-gin/internal/domain/entity"
	"backend-gin/pkg/pagination"
)

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
}

type UpdateUserRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListUsersRequest struct {
	pagination.Request
}

type ListUsersResponse struct {
	Users []*UserResponse    `json:"users"`
	Meta  *pagination.Result `json:"meta"`
}

func ToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToUserResponses(users []*entity.User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(user)
	}
	return responses
}
