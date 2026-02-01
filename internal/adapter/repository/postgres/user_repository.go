package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-gin/db/sqlc"
	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
)

type userRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	role := user.Role
	if role == "" {
		role = entity.RoleUser
	}

	result, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		Role:         string(role),
	})
	if err != nil {
		return nil, err
	}
	return toUserEntity(&result), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	result, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserEntity(&result), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	result, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toUserEntity(&result), nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int32) ([]*entity.User, error) {
	results, err := r.queries.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	return toUserEntities(results), nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	result, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:   user.ID,
		Name: user.Name,
	})
	if err != nil {
		return nil, err
	}
	return toUserEntity(&result), nil
}

func (r *userRepository) UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role) (*entity.User, error) {
	result, err := r.queries.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
		ID:   id,
		Role: string(role),
	})
	if err != nil {
		return nil, err
	}
	return toUserEntity(&result), nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, id)
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

func toUserEntity(u *sqlc.User) *entity.User {
	return &entity.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		Role:         entity.ParseRole(u.Role),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func toUserEntities(users []sqlc.User) []*entity.User {
	entities := make([]*entity.User, len(users))
	for i, u := range users {
		entities[i] = toUserEntity(&u)
	}
	return entities
}
