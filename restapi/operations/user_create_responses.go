// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "db-forum/models"
)

// UserCreateCreatedCode is the HTTP code returned for type UserCreateCreated
const UserCreateCreatedCode int = 201

/*UserCreateCreated Пользователь успешно создан.
Возвращает данные созданного пользователя.


swagger:response userCreateCreated
*/
type UserCreateCreated struct {

	/*
	  In: Body
	*/
	Payload *models.User `json:"body,omitempty"`
}

// NewUserCreateCreated creates UserCreateCreated with default headers values
func NewUserCreateCreated() *UserCreateCreated {

	return &UserCreateCreated{}
}

// WithPayload adds the payload to the user create created response
func (o *UserCreateCreated) WithPayload(payload *models.User) *UserCreateCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the user create created response
func (o *UserCreateCreated) SetPayload(payload *models.User) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UserCreateCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UserCreateConflictCode is the HTTP code returned for type UserCreateConflict
const UserCreateConflictCode int = 409

/*UserCreateConflict Пользователь уже присутсвует в базе данных.
Возвращает данные ранее созданных пользователей с тем же nickname-ом иои email-ом.


swagger:response userCreateConflict
*/
type UserCreateConflict struct {

	/*
	  In: Body
	*/
	Payload models.Users `json:"body,omitempty"`
}

// NewUserCreateConflict creates UserCreateConflict with default headers values
func NewUserCreateConflict() *UserCreateConflict {

	return &UserCreateConflict{}
}

// WithPayload adds the payload to the user create conflict response
func (o *UserCreateConflict) WithPayload(payload models.Users) *UserCreateConflict {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the user create conflict response
func (o *UserCreateConflict) SetPayload(payload models.Users) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UserCreateConflict) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(409)
	payload := o.Payload
	if payload == nil {
		payload = make(models.Users, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}
