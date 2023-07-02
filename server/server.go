package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	jwt_http2 "github.com/Moranilt/jwt-http2/auth"
	"github.com/Moranilt/jwt-http2/config"
	"github.com/Moranilt/jwt-http2/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	ERROR_StoreTokenToRedis          = "cannot store token to redis: %v"
	ERROR_MakeAccessToken            = "make access token: %v"
	ERROR_MakeRefreshToken           = "make refresh token: %v"
	ERROR_RefreshTokenNotFound       = "refresh token not found"
	ERROR_TokenNotFound              = "token not found"
	ERROR_ProvideAnyField            = "provide any field"
	ERROR_CannotDeleteTokenFromRedis = "cannot delete token from redis. Error: %v"
)

type Server struct {
	jwt_http2.UnimplementedAuthenticationServer
	log         *logger.Logger
	config      *config.AppConfig[time.Duration]
	redis       *redis.Client
	publicCert  []byte
	privateCert []byte
}

type UserClaims = map[string]string

type AccessClaims struct {
	UUID       string     `json:"session"`
	UserClaims UserClaims `json:"user_claims"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	AccessUUID  string     `json:"access_uuid"`
	RefreshUUID string     `json:"refresh_uuid"`
	UserClaims  UserClaims `json:"user_claims"`
	jwt.RegisteredClaims
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(
	log *logger.Logger,
	config *config.AppConfig[time.Duration],
	r *redis.Client,
	public, private []byte,
) *Server {
	return &Server{
		log:         log,
		config:      config,
		redis:       r,
		publicCert:  public,
		privateCert: private,
	}
}

func (s *Server) CreateTokens(ctx context.Context, req *jwt_http2.CreateTokensRequest) (*jwt_http2.CreateTokensResponse, error) {
	log := s.log.WithRequestInfo(ctx)
	log.WithFields(logrus.Fields{
		"req": req,
	}).Info()

	tokens, err := s.makeNewTokens(ctx, req.UserId, req.UserClaims)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &jwt_http2.CreateTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) RefreshTokens(ctx context.Context, req *jwt_http2.RefreshTokensRequest) (*jwt_http2.RefreshTokenResponse, error) {
	log := s.log.WithRequestInfo(ctx)
	log.WithFields(logrus.Fields{
		"req": req,
	}).Info()

	claims, err := s.parseRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Error("parse refresh token: ", err)
		return nil, err
	}

	userId, err := s.redis.Get(ctx, claims.RefreshUUID).Result()
	if err != nil {
		if err == redis.Nil {
			log.Error(ERROR_RefreshTokenNotFound)
			return nil, errors.New(ERROR_RefreshTokenNotFound)
		}
		log.Error("redis: ", err)
		return nil, err
	}

	err = s.redis.Del(ctx, claims.RefreshUUID, claims.AccessUUID).Err()
	if err != nil {
		log.Error("redis: ", err)
		return nil, err
	}

	newTokens, err := s.makeNewTokens(ctx, userId, claims.UserClaims)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &jwt_http2.RefreshTokenResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
	}, nil
}

func (s *Server) GetUserId(ctx context.Context, req *jwt_http2.GetUserIdRequest) (*jwt_http2.GetUserIdResponse, error) {
	log := s.log.WithRequestInfo(ctx)
	log.WithFields(logrus.Fields{
		"req": req,
	}).Info()

	claims, err := s.parseAccessToken(ctx, req.AccessToken)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	userId, err := s.redis.Get(ctx, claims.UUID).Result()
	if err != nil {
		if err == redis.Nil {
			log.Error(ERROR_TokenNotFound)
			return nil, errors.New(ERROR_TokenNotFound)
		}
		log.Error("redis: ", err)
		return nil, err
	}

	return &jwt_http2.GetUserIdResponse{
		UserId: userId,
	}, nil
}

func (s *Server) CheckTokenExistence(ctx context.Context, req *jwt_http2.CheckTokenExistenceRequest) (*jwt_http2.CheckTokenExistenceResponse, error) {
	log := s.log.WithRequestInfo(ctx)
	log.WithFields(logrus.Fields{
		"req": req,
	}).Info()

	if req.GetAccessToken() == "" && req.GetRefreshToken() == "" {
		log.Error(ERROR_ProvideAnyField)
		return nil, fmt.Errorf(ERROR_ProvideAnyField)
	}

	response := new(jwt_http2.CheckTokenExistenceResponse)

	if req.GetAccessToken() != "" {
		claims, err := s.parseAccessToken(ctx, req.GetAccessToken())
		if err != nil {
			log.Error(err)
			return nil, err
		}

		result, err := s.redis.Exists(ctx, claims.UUID).Result()
		if err != nil {
			log.Error(err)
			return nil, err
		}

		access := result == 1
		response.AccessToken = &access
	}

	if req.GetRefreshToken() != "" {
		claims, err := s.parseRefreshToken(ctx, req.GetRefreshToken())
		if err != nil {
			log.Error(err)
			return nil, err
		}

		result, err := s.redis.Exists(ctx, claims.RefreshUUID).Result()
		if err != nil {
			log.Error(err)
			return nil, err
		}

		refresh := result == 1
		response.RefreshToken = &refresh
	}

	return response, nil
}

func (s *Server) RevokeTokens(ctx context.Context, req *jwt_http2.RevokeTokensRequest) (*jwt_http2.RevokeTokensResponse, error) {
	log := s.log.WithRequestInfo(ctx)
	log.WithFields(logrus.Fields{
		"req": req,
	}).Info()

	claims, err := s.parseRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = s.redis.Del(ctx, claims.RefreshUUID).Err()
	if err != nil {
		log.Errorf(ERROR_CannotDeleteTokenFromRedis, err)
		return nil, fmt.Errorf(ERROR_CannotDeleteTokenFromRedis, err)
	}

	err = s.redis.Del(ctx, claims.AccessUUID).Err()
	if err != nil {
		log.Errorf(ERROR_CannotDeleteTokenFromRedis, err)
		return nil, fmt.Errorf(ERROR_CannotDeleteTokenFromRedis, err)
	}

	return &jwt_http2.RevokeTokensResponse{
		Revoked: true,
	}, nil
}

func (s *Server) makeAccessToken(ctx context.Context, uuid string, uc UserClaims, exp time.Time) (string, error) {
	claims := AccessClaims{
		UUID:       uuid,
		UserClaims: uc,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   s.config.Subject,
			Audience:  s.config.Audience,
			ID:        uuid,
		},
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(s.privateCert)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	access_token, err := token.SignedString(key)
	if err != nil {
		return "", errors.New("cannot create new token. Error: " + err.Error())
	}

	return access_token, nil
}

func (s *Server) makeRefreshToken(ctx context.Context, accessUUID string, refreshUUID string, uc UserClaims, refreshExp time.Time) (string, error) {
	claims := RefreshClaims{
		AccessUUID:  accessUUID,
		RefreshUUID: refreshUUID,
		UserClaims:  uc,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   s.config.Subject,
			Audience:  s.config.Audience,
			ID:        refreshUUID,
		},
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(s.privateCert)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	refresh_token, err := token.SignedString(key)
	if err != nil {
		return "", errors.New("cannot create new token. Error: " + err.Error())
	}

	return refresh_token, nil
}

func (s *Server) makeJwtOptions(options ...jwt.ParserOption) []jwt.ParserOption {
	var o []jwt.ParserOption
	o = append(o, options...)
	o = append(o, jwt.WithSubject(s.config.Subject), jwt.WithIssuer(s.config.Issuer))
	for _, aud := range s.config.Audience {
		o = append(o, jwt.WithAudience(aud))
	}
	return o
}

func (s *Server) parseRefreshToken(ctx context.Context, refreshToken string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &RefreshClaims{}, func(t *jwt.Token) (any, error) {
		key, err := jwt.ParseRSAPublicKeyFromPEM(s.publicCert)
		if err != nil {
			return nil, err
		}
		return key, nil
	}, s.makeJwtOptions()...)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("not valid token claims")
	}
}

func (s *Server) parseAccessToken(ctx context.Context, refreshToken string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		key, err := jwt.ParseRSAPublicKeyFromPEM(s.publicCert)
		if err != nil {
			return nil, err
		}
		return key, nil
	}, s.makeJwtOptions()...)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("not valid token claims")
	}
}

func (s *Server) makeNewTokens(ctx context.Context, userId string, userClaims UserClaims) (*AuthTokens, error) {
	now := time.Now()
	accessUUID := uuid.NewString()
	accessExp := now.Add(s.config.TTL.Access)

	refreshUUID := uuid.NewString()
	refreshExp := now.Add(s.config.TTL.Refresh)

	access_token, err := s.makeAccessToken(ctx, accessUUID, userClaims, accessExp)
	if err != nil {
		return nil, fmt.Errorf(ERROR_MakeAccessToken, err)
	}

	refresh_token, err := s.makeRefreshToken(ctx, accessUUID, refreshUUID, userClaims, refreshExp)
	if err != nil {
		return nil, fmt.Errorf(ERROR_MakeRefreshToken, err)
	}

	err = s.redis.Set(ctx, accessUUID, userId, time.Until(accessExp)).Err()
	if err != nil {
		return nil, fmt.Errorf(ERROR_StoreTokenToRedis, err)
	}

	err = s.redis.Set(ctx, refreshUUID, userId, time.Until(refreshExp)).Err()
	if err != nil {
		return nil, fmt.Errorf(ERROR_StoreTokenToRedis, err)
	}

	return &AuthTokens{
		AccessToken:  access_token,
		RefreshToken: refresh_token,
	}, nil
}
