package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/dgrijalva/jwt-go"
	jwtreq "github.com/dgrijalva/jwt-go/request"
	"github.com/julienschmidt/httprouter"
)

// AuthSuccessType contains TypeInfo for the AuthSuccess type
var AuthSuccessType = &runtime.TypeInfo{
	Kind:        "auth-success",
	Constructor: func() runtime.Object { return &AuthSuccess{} },
}

// AuthSuccess represents successful authentication
type AuthSuccess struct {
	runtime.TypeKind `yaml:",inline"`
	Token            string
}

// AuthRequestType contains TypeInfo for the AuthRequest type
var AuthRequestType = &runtime.TypeInfo{
	Kind:        "auth-request",
	Constructor: func() runtime.Object { return &AuthRequest{} },
}

// AuthRequest represents authentication request
type AuthRequest struct {
	runtime.TypeKind `yaml:",inline"`
	Username         string
	Password         string
}

func (api *coreAPI) handleLogin(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	authReq, ok := api.contentType.ReadOne(request).(*AuthRequest)
	if !ok {
		panic(fmt.Sprintf("Unexpected object received: %v", authReq))
	}

	user, err := api.externalData.UserLoader.Authenticate(authReq.Username, authReq.Password)
	if err != nil {
		serverErr := NewServerError(fmt.Sprintf("Authentication error: %s", err))
		api.contentType.WriteOne(writer, request, serverErr)
	} else {
		api.contentType.WriteOne(writer, request, &AuthSuccess{
			TypeKind: AuthSuccessType.GetTypeKind(),
			Token:    api.newToken(user),
		})
	}
}

// Claims represent Aptomi JWT Claims
type Claims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

// Valid checks if claims are valid
func (claims Claims) Valid() error {
	if len(claims.Name) == 0 {
		return fmt.Errorf("token should contain non-empty username")
	}

	return claims.StandardClaims.Valid()
}

func (api *coreAPI) newToken(user *lang.User) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Name: user.Name,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
		},
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(api.secret))
	if err != nil {
		panic(fmt.Errorf("error while signing token: %s", err))
	}

	return tokenString
}

func (api *coreAPI) auth(handle httprouter.Handle) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		err := api.checkToken(request)
		if err != nil {
			authErr := NewServerError(fmt.Sprintf("Authentication error: %s", err))
			api.contentType.WriteOneWithStatus(writer, request, authErr, http.StatusUnauthorized)
			return
		}

		handle(writer, request, params)
	}
}

// The key type is unexported to prevent collisions with context keys defined in other packages
type key int

const (
	// ctxUserKey is the context key for user
	ctxUserKey key = iota
)

func (api *coreAPI) checkToken(request *http.Request) error {
	token, err := jwtreq.ParseFromRequestWithClaims(request, jwtreq.AuthorizationHeaderExtractor, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(api.secret), nil
		})
	if err != nil {
		return err
	}
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return fmt.Errorf("unexpected token signing method: %s", token.Header["alg"])
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return fmt.Errorf("unexpected token claims, can't be casted to *Claims: %s", token.Claims)
	}

	user := api.externalData.UserLoader.LoadUserByName(claims.Name)
	if user == nil {
		return fmt.Errorf("token refers to non-existing user: %s", claims.Name)
	}

	// registry user into the request
	newRequest := request.WithContext(context.WithValue(request.Context(), ctxUserKey, user))
	*request = *newRequest

	return nil
}

func (api *coreAPI) getUserOptional(request *http.Request) *lang.User {
	val := request.Context().Value(ctxUserKey)
	if val == nil {
		return nil
	}
	if user, ok := val.(*lang.User); ok {
		return user
	}

	return nil
}

func (api *coreAPI) getUserRequired(request *http.Request) *lang.User {
	user := api.getUserOptional(request)
	if user == nil {
		panic("unauthorized or user couldn't be loaded")
	}

	return user
}
