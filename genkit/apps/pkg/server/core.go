package server

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/code"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/middleware"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/mingyuans/errors"
	"reflect"
)

const (
	// invalidStatusCode means the code isn't setup, please set it.
	invalidStatusCode = 0
)

type Meta struct {
	// Business error code. Please check the code with our docs.
	//
	// Example: 0
	Code int `json:"code"`
	// The detail message of the error.
	//
	// Example: The user existed.
	Message string `json:"message"`
	// The other messages. But most of this will be empty.
	//
	// Example: []
	Errors []string `json:"errors"`
	// The request id of the request.
	//
	// Example: 67575010234d4f9f9adaca7c26e7e709
	RequestId string `json:"request_id"`
}

type Response struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

type builder struct {
	err        error
	context    *gin.Context
	statusCode int
	*Response
}

func NewRestfulResponseBuilder(c *gin.Context) *builder {
	return &builder{
		statusCode: invalidStatusCode,
		context:    c,
		Response: &Response{
			Data: nil,
			Meta: Meta{
				Code: invalidStatusCode,
			},
		},
	}
}

func (b *builder) Meta(meta Meta) *builder {
	b.Response.Meta = meta
	return b
}

func (b *builder) Data(data interface{}) *builder {
	b.Response.Data = data
	return b
}

func (b *builder) Error(err error) *builder {
	b.err = err
	return b
}

func (b *builder) StatusCode(statusCode int) *builder {
	b.statusCode = statusCode
	return b
}

func (b *builder) Build() (int, Response) {
	if b.err != nil {
		return b.buildErrorResponse()
	}
	return b.buildSuccessResponse()
}

func (b *builder) getErrorMessages(err error) []string {
	var messages []string
	if err == nil {
		return messages
	}

	messages = append(messages, err.Error())

	cause := errors.Cause(b.err)
	if cause == nil ||
		//通过类型判断是否是同一个error
		errors.Is(cause, b.err) ||
		//通过指针地址判断是否是同一个error,比如 validator.ValidationErrors 是一个 []error, 需要通过指针地址来判断
		reflect.ValueOf(cause).Pointer() == reflect.ValueOf(err).Pointer() {
		return messages
	}

	messages = append(messages, cause.Error())
	return messages
}

func (b *builder) buildErrorResponse() (int, Response) {
	coder := errors.ParseCoder(b.err)
	statusCode := b.buildStatusCode(coder.HTTPStatus())
	b.Response.Meta.Code = int(coder.Code())
	b.Response.Meta.Message = coder.String()
	requestId := b.context.GetString(middleware.XRequestIDKey)
	b.Response.Meta.RequestId = requestId
	if !errors.IsCode(b.err, code.Success) {
		log.Errorf("%#+v", b.err)
		b.Response.Meta.Errors = b.getErrorMessages(b.err)
	}
	return statusCode, *b.Response
}

func (b *builder) buildStatusCode(statusCode int) int {
	if b.statusCode != invalidStatusCode {
		return b.statusCode
	}
	return statusCode
}

func (b *builder) buildSuccessResponse() (int, Response) {
	b.err = errors.WithCode(code.Success, "")
	statusCode, response := b.buildErrorResponse()
	// we don't fill meta.errors when the request is success.
	response.Meta.Errors = nil
	return statusCode, response
}

func (b *builder) SendJSON() {
	statusCode, response := b.Build()
	b.context.JSON(statusCode, response)
}
