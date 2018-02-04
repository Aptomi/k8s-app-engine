package api

import (
	"context"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/dgrijalva/jwt-go"
	jwtreq "github.com/dgrijalva/jwt-go/request"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

// AuthSuccessObject contains Info for the AuthSuccess type
var AuthSuccessObject = &runtime.Info{
	Kind:        "auth-success",
	Constructor: func() runtime.Object { return &AuthSuccess{} },
}

// AuthSuccess represents successful authentication
type AuthSuccess struct {
	runtime.TypeKind `yaml:",inline"`
	Token            string
}

func (api *coreAPI) handleLogin(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	username := request.PostFormValue("username")
	password := request.PostFormValue("password")
	user, err := api.externalData.UserLoader.Authenticate(username, password)
	if err != nil {
		serverErr := NewServerError(fmt.Sprintf("Authentication error: %s", err))
		api.contentType.WriteOne(writer, request, serverErr)
	} else {
		api.contentType.WriteOne(writer, request, &AuthSuccess{
			AuthSuccessObject.GetTypeKind(),
			api.newToken(user),
		})
	}
}

type Claims struct {
	Name        string `json:"name"`
	DomainAdmin bool   `json:"admin,omitempty"`
	jwt.StandardClaims
}

func (claims Claims) Valid() error {
	if len(claims.Name) == 0 {
		return fmt.Errorf("token should contain non-empty username")
	}

	return claims.StandardClaims.Valid()
}

func (api *coreAPI) newToken(user *lang.User) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Name:        user.Name,
		DomainAdmin: user.DomainAdmin,
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
	return api.handleAuth(handle, false)
}

func (api *coreAPI) admin(handle httprouter.Handle) httprouter.Handle {
	return api.handleAuth(handle, true)
}

func (api *coreAPI) handleAuth(handle httprouter.Handle, admin bool) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		err := api.checkToken(request, admin)
		if err != nil {
			authErr := NewServerError(fmt.Sprintf("Authentication error: %s", err))
			api.contentType.WriteOneWithStatus(writer, request, authErr, http.StatusUnauthorized)
			return
		}

		handle(writer, request, params)
	}
}

const (
	ctxUserProperty = "user"
)

func (api *coreAPI) checkToken(request *http.Request, admin bool) error {
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

	claims := token.Claims.(*Claims)
	user := api.externalData.UserLoader.LoadUserByName(claims.Name)
	if user == nil {
		return fmt.Errorf("token refers to non-existing user: %s", claims.Name)
	}
	if user.DomainAdmin != claims.DomainAdmin {
		return fmt.Errorf("token contains incorrect admin status: %s", claims.DomainAdmin)
	}

	if admin && !user.DomainAdmin {
		return fmt.Errorf("admin privileges required")
	}

	// store user into the request
	newRequest := request.WithContext(context.WithValue(request.Context(), ctxUserProperty, user))
	*request = *newRequest

	return nil
}

func (api *coreAPI) getUserOptional(request *http.Request) *lang.User {
	val := request.Context().Value(ctxUserProperty)
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
