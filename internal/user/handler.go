package user

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"learn/internal/apperror"
	"learn/internal/handlers"
	"learn/pkg/logging"
	"net/http"
	"time"
)

const (
	usersURL               = "/users"
	userURL                = "/users/:uuid"
	createUserURL          = "/users/:email/:username/:password"
	updateUserURL          = "/users/:uuid/:email/:username/:password"
	partiallyUpdateUserURL = "/users/:uuid/:email/:username/:password"
)

var _ handlers.Handler = &handler{}

type handler struct {
	logger *logging.Logger
	stg    Storage
}

func NewHandler(logger *logging.Logger, storage Storage) handlers.Handler {
	return &handler{logger: logger, stg: storage}
}

func (h *handler) Register(router *httprouter.Router) {
	router.GET(usersURL, apperror.Middleware(h.GetList))
	router.POST(createUserURL, apperror.Middleware(h.CreateUser))
	router.GET(userURL, apperror.Middleware(h.GetUserByUUID))
	router.PUT(updateUserURL, apperror.Middleware(h.UpdateUser))
	router.PATCH(partiallyUpdateUserURL, apperror.Middleware(h.PartiallyUpdateUser))
	router.DELETE(userURL, apperror.Middleware(h.DeleteUser))
}

func (h *handler) GetList(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {

	users, err := h.stg.FindAll(context.Background())
	h.logger.Trace(users)
	if err != nil {
		return apperror.ErrNotFound
	}

	data, err := json.Marshal(users)
	if err != nil {
		h.logger.Info(err)
		return apperror.NewAppError(err, "error marshaling", err.Error(), "E-0990")
	}

	w.Write(data)
	w.WriteHeader(200)

	return nil
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {

	user := CreateUserDTO{
		Email:    params.ByName("email"),
		Username: params.ByName("username"),
		Password: params.ByName("password"),
	}

	timeout, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*10))
	defer cancelFunc()

	_, err := h.stg.Create(timeout, User{
		ID:           "",
		Email:        user.Email,
		Username:     user.Username,
		PasswordHash: user.Password,
	})

	if err != nil {
		return apperror.NewAppError(err, fmt.Sprintf("error create user with name: %s", user.Username), err.Error(), "E-0991")
	}

	h.logger.Tracef("Created user %v", user)
	userCreateBytes, err := json.Marshal(user)
	h.logger.Trace(userCreateBytes)
	if err != nil {
		return apperror.NewAppError(err, "error marshaling", err.Error(), "E-0990")
	}

	w.Write(userCreateBytes)
	w.WriteHeader(201)

	return nil
}

func (h *handler) GetUserByUUID(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	oId, err := primitive.ObjectIDFromHex(params.ByName("uuid"))

	user, err := h.stg.FindOne(context.Background(), oId.Hex())
	if err != nil {
		h.logger.Info(err)
		return apperror.NewAppError(err, fmt.Sprintf("error find one user by id: %s", oId.Hex()), err.Error(), "E-0990")
	}

	data, err := json.Marshal(user)
	if err != nil {
		h.logger.Info(err)
		return apperror.NewAppError(err, fmt.Sprintf("error marshaling user by id: %s", oId.Hex()), err.Error(), "E-0990")
	}

	w.Write(data)
	w.WriteHeader(200)

	return nil
}

func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {

	oId, err := primitive.ObjectIDFromHex(params.ByName("uuid"))
	if err != nil {
		return apperror.NewAppError(err, fmt.Sprintf("failed to convert user ID to ObjectId. ID=%s", oId.Hex()), err.Error(), "E-0992")
	}

	user, err := h.stg.FindOne(context.Background(), oId.Hex())
	h.logger.Trace(user)
	if err != nil {
		return err
	}

	user.Email = params.ByName("email")
	user.Username = params.ByName("username")
	user.PasswordHash = params.ByName("password")

	errUpdate := h.stg.Update(context.Background(), user)
	if errUpdate != nil {
		return apperror.NewAppError(err, fmt.Sprintf("error update user id: %s", oId.Hex()), err.Error(), "E-0993")
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return apperror.NewAppError(err, fmt.Sprintf("error marshaling user by id: %s", oId.Hex()), err.Error(), "E-0990")
	}

	w.Write(userBytes)
	w.WriteHeader(200)
	return nil
}

func (h *handler) PartiallyUpdateUser(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {

	oId, err := primitive.ObjectIDFromHex(params.ByName("uuid"))
	if err != nil {
		return apperror.NewAppError(err, fmt.Sprintf("failed to convert user ID to ObjectId. ID=%s", oId.Hex()), err.Error(), "E-0992")
	}

	user, err := h.stg.FindOne(context.Background(), oId.Hex())
	h.logger.Trace(user)
	if err != nil {
		return err
	}

	email := params.ByName("email")
	username := params.ByName("username")
	password := params.ByName("password")

	if email != "pass" {
		user.Email = email
	}
	if username != "pass" {
		user.Username = username
	}
	if password != "pass" {
		user.PasswordHash = password
	}

	errUpdate := h.stg.Update(context.Background(), user)
	if errUpdate != nil {
		return apperror.NewAppError(err, fmt.Sprintf("error partially update user id: %s", oId.Hex()), err.Error(), "E-0993")
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return apperror.NewAppError(err, fmt.Sprintf("error marshaling user by id: %s", oId.Hex()), err.Error(), "E-0990")
	}

	w.Write(userBytes)
	w.WriteHeader(200)

	return nil
}

func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	oId, errObjFromHex := primitive.ObjectIDFromHex(params.ByName("uuid"))

	if errObjFromHex != nil {
		return apperror.NewAppError(errObjFromHex, fmt.Sprintf("failed to convert user ID to ObjectId. ID=%s", oId.Hex()), errObjFromHex.Error(), "E-0992")
	}

	errDelete := h.stg.Delete(context.Background(), oId.Hex())
	if errDelete != nil {
		return apperror.ErrDelete
	}

	w.Write([]byte(fmt.Sprintf("Пользователь с ID: %s успешно удален!", oId)))

	return nil
}
