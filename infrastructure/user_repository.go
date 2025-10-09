package infrastructure

import (
	"context"
	"fmt"

	"github.com/knetic0/production-ready-go-cqrs/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type UserRepositoryAdapter struct {
	db *gorm.DB
}

func NewUserRepositoryAdapter(db *gorm.DB) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{db: db}
}

func (r *UserRepositoryAdapter) Create(ctx context.Context, user *domain.User) error {
	tracer := otel.Tracer("microservice-go/repository.user")
	ctx, span := tracer.Start(ctx, "UserRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("user.id", user.Id),
		attribute.String("user.email", user.Email),
	)

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db.create failed")
		return err
	}

	span.SetStatus(codes.Ok, "created")
	return nil
}

func (r *UserRepositoryAdapter) Get(ctx context.Context, id string) (*domain.User, error) {
	tracer := otel.Tracer("microservice-go/repository.user")
	ctx, span := tracer.Start(ctx, "UserRepository.Get")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("user.id", id),
	)

	u, err := r.getByField(ctx, "id", id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "found")
	return u, nil
}

func (r *UserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	tracer := otel.Tracer("microservice-go/repository.user")
	ctx, span := tracer.Start(ctx, "UserRepository.GetByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("user.email", email),
	)

	u, err := r.getByField(ctx, "email", email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "getByEmail failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "found")
	return u, nil
}

func (r *UserRepositoryAdapter) List(ctx context.Context) ([]domain.User, error) {
	tracer := otel.Tracer("microservice-go/repository.user")
	ctx, span := tracer.Start(ctx, "UserRepository.List")
	defer span.End()

	span.SetAttributes(attribute.String("db.system", "postgresql"))

	var users []domain.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "list failed")
		return nil, err
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))
	span.SetStatus(codes.Ok, "ok")
	return users, nil
}

func (r *UserRepositoryAdapter) getByField(ctx context.Context, field string, value any) (*domain.User, error) {
	tracer := otel.Tracer("microservice-go/repository.user")
	ctx, span := tracer.Start(ctx, "UserRepository.getByField")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("query.field", field),
		attribute.String("query.value", fmt.Sprintf("%v", value)),
	)

	var u domain.User
	if err := r.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", field), value).Take(&u).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "getByField failed")
		return nil, err
	}

	span.SetAttributes(attribute.String("user.id", u.Id), attribute.String("user.email", u.Email))
	span.SetStatus(codes.Ok, "found")
	return &u, nil
}
