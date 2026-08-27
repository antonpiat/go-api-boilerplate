package user

import (
	"context"
	"errors"

	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already registered")
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, page httpx.Pagination) ([]User, int64, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, u *User) error {
	err := r.db.WithContext(ctx).Create(u).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailTaken
	}
	return err
}

func (r *GormRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Take(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *GormRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("email = ?", email).Take(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *GormRepository) List(ctx context.Context, page httpx.Pagination) ([]User, int64, error) {
	var (
		users []User
		total int64
	)
	q := r.db.WithContext(ctx).Model(&User{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []User{}, 0, nil
	}
	err := q.Session(&gorm.Session{}).
		Select("id", "email", "role", "created_at", "updated_at").
		Order("created_at DESC").
		Limit(page.PerPage).
		Offset(page.Offset()).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *GormRepository) Update(ctx context.Context, u *User) error {
	err := r.db.WithContext(ctx).Save(u).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailTaken
	}
	return err
}

func (r *GormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
