package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const requestIDKey = "request_id"

type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func Proto(c *gin.Context, status int, msg proto.Message) {
	if msg == nil {
		c.Status(status)
		return
	}

	data, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}.Marshal(msg)
	if err != nil {
		Error(c, http.StatusInternalServerError, "internal", "failed to encode response")
		return
	}

	c.Data(status, "application/json; charset=utf-8", data)
}

func Error(c *gin.Context, status int, code string, message string) {
	c.AbortWithStatusJSON(status, ErrorBody{
		Error: ErrorPayload{
			Code:      code,
			Message:   message,
			RequestID: RequestID(c),
		},
	})
}

func FromGRPCError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		Error(c, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	httpStatus, code := grpcCodeToHTTP(st.Code())
	Error(c, httpStatus, code, st.Message())
}

func RequestID(c *gin.Context) string {
	if value, ok := c.Get(requestIDKey); ok {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return ""
}

func SetRequestID(c *gin.Context, requestID string) {
	c.Set(requestIDKey, requestID)
}

func grpcCodeToHTTP(code codes.Code) (int, string) {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest, "invalid_argument"
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "unauthenticated"
	case codes.PermissionDenied:
		return http.StatusForbidden, "permission_denied"
	case codes.NotFound:
		return http.StatusNotFound, "not_found"
	case codes.AlreadyExists:
		return http.StatusConflict, "already_exists"
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed, "failed_precondition"
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, "resource_exhausted"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "deadline_exceeded"
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "unavailable"
	case codes.Canceled:
		return http.StatusRequestTimeout, "canceled"
	default:
		return http.StatusInternalServerError, "internal"
	}
}
