package domain

import (
	"authX/utils"
	"authX/utils/config"
	"authX/utils/constants"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type DomainService struct {
	logger     *zap.Logger
	postgreDB  PostgreRepo
	redisDB    RedisRepo
	hasher     *utils.BcryptHasher
	jwtManager *utils.JWTManager
	config     *config.Config
}

func NewDomainService(logger *zap.Logger, config *config.Config, postgreRepo PostgreRepo, hasher *utils.BcryptHasher, jwtManager *utils.JWTManager, redisRepo RedisRepo) *DomainService {
	return &DomainService{
		logger:     logger,
		postgreDB:  postgreRepo,
		redisDB:    redisRepo,
		hasher:     hasher,
		jwtManager: jwtManager,
		config:     config,
	}
}

func (s *DomainService) RegisterUser(ctx context.Context, dto RegisterUserDto) error {
	existingUser, err := s.postgreDB.GetByEmail(ctx, dto.Email)
	if err != nil {
		s.logger.Error("Failed to check email", zap.Error(err))
		return fmt.Errorf(constants.Internal)
	}
	if existingUser != nil {
		return fmt.Errorf("email %s already registered", dto.Email)
	}

	hashedPassword, err := s.hasher.Hash(dto.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return fmt.Errorf(constants.Internal)
	}

	user := &User{
		Email:    dto.Email,
		Role:     constants.UserRoleBase,
		Password: hashedPassword,
	}

	if err := s.postgreDB.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf(constants.Internal)
	}

	s.logger.Info("User registered",
		zap.String("email", user.Email),
	)

	return nil
}

func (s *DomainService) Login(ctx context.Context, dto LoginUserDto) (*UserLoginResp, error) {
	user, err := s.postgreDB.GetByEmail(ctx, dto.Email)
	if err != nil {
		s.logger.Error("Failed to find user", zap.Error(err))
		return nil, fmt.Errorf(constants.Internal)
	}
	if user == nil {
		return nil, fmt.Errorf(constants.NotFound)
	}

	if !s.hasher.Check(dto.Password, user.Password) {
		return nil, fmt.Errorf(constants.PasswordNotWalid)
	}

	oldToken, err := s.redisDB.GetUserToken(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get old token", zap.Error(err))
	}

	if oldToken != "" {
		oldClaims, err := s.jwtManager.Validate(oldToken)
		if err == nil {
			ttl := time.Until(oldClaims.ExpiresAt.Time)
			if ttl > 0 {
				if err := s.redisDB.AddToBlacklist(ctx, oldClaims.TokenID, ttl); err != nil {
					s.logger.Error("Failed to blacklist old token", zap.Error(err))
				}
			}
		}
	}

	role := ""
	if user.Role != "" {
		role = user.Role
	}

	newToken, _, err := s.jwtManager.Generate(user.ID, user.Email, role)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	tokenTTL := 15 * time.Minute
	if err := s.redisDB.SaveUserToken(ctx, user.ID, newToken, tokenTTL); err != nil {
		s.logger.Error("Failed to save token in Redis", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	s.logger.Info("User logged in",
		zap.String("email", user.Email),
	)

	return &UserLoginResp{
		User:  *user,
		Token: newToken,
	}, nil
}

func (s *DomainService) ValidateToken(ctx context.Context, tokenString string) (*ValidateResponse, error) {
	claims, err := s.jwtManager.Validate(tokenString)
	if err != nil {
		return nil, fmt.Errorf(constants.TokenNotValid)
	}

	isBlacklisted, err := s.redisDB.IsBlacklisted(ctx, claims.TokenID)
	if err != nil {
		s.logger.Error("Failed to check blacklist", zap.Error(err))
		return nil, fmt.Errorf(constants.Internal)
	}
	if isBlacklisted {
		return nil, fmt.Errorf(constants.TokenNotValid)
	}

	return &ValidateResponse{
		Valid:  true,
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (s *DomainService) GetRoles() []string {
	return constants.BaseRoles
}