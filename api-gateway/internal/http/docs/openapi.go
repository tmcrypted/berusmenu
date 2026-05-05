package docs

import "encoding/json"

func OpenAPIJSON() ([]byte, error) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Berusmenu API Gateway",
			"version":     "0.1.0",
			"description": "Public HTTP entrypoint for Berusmenu. Gateway stays thin: it validates auth context, blocks unverified users outside the whitelist, forwards identity via gRPC metadata, and maps REST to downstream gRPC services.",
		},
		"servers": []map[string]any{
			{"url": "/"},
		},
		"tags": []map[string]any{
			{"name": "Health"},
			{"name": "Auth"},
			{"name": "Profile"},
			{"name": "PVZ"},
			{"name": "Moderation"},
			{"name": "Vacancies"},
			{"name": "Responses"},
			{"name": "Assignments"},
			{"name": "Reviews"},
			{"name": "Chat"},
			{"name": "WebSocket"},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
			"schemas": schemas(),
		},
		"paths": paths(),
	}

	return json.MarshalIndent(spec, "", "  ")
}

func schemas() map[string]any {
	return map[string]any{
		"HealthStatus": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{"type": "string", "example": "ok"},
			},
			"required": []string{"status"},
		},
		"ErrorResponse": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"error": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"code":       map[string]any{"type": "string", "example": "invalid_argument"},
						"message":    map[string]any{"type": "string", "example": "validation failed"},
						"request_id": map[string]any{"type": "string", "example": "req-123"},
					},
					"required": []string{"code", "message", "request_id"},
				},
			},
			"required": []string{"error"},
		},
		"UserRole": map[string]any{
			"type": "string",
			"enum": []string{
				"USER_ROLE_UNSPECIFIED",
				"USER_ROLE_OWNER",
				"USER_ROLE_WORKER",
				"USER_ROLE_MODERATOR",
			},
		},
		"User": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":                map[string]any{"type": "integer", "format": "int64", "example": 7},
				"email":             map[string]any{"type": "string", "format": "email", "example": "worker@example.com"},
				"email_verified_at": map[string]any{"type": "string", "format": "date-time", "nullable": true},
				"role":              map[string]any{"$ref": "#/components/schemas/UserRole"},
				"full_name":         map[string]any{"type": "string", "example": "Worker User"},
				"phone":             map[string]any{"type": "string", "nullable": true, "example": "+79990001122"},
				"avatar_url":        map[string]any{"type": "string", "format": "uri", "nullable": true},
				"created_at":        map[string]any{"type": "string", "format": "date-time"},
				"updated_at":        map[string]any{"type": "string", "format": "date-time"},
			},
			"required": []string{"id", "email", "role", "full_name", "created_at", "updated_at"},
		},
		"AuthTokens": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"access_token":      map[string]any{"type": "string"},
				"refresh_token":     map[string]any{"type": "string"},
				"access_expires_at": map[string]any{"type": "string", "format": "date-time"},
			},
			"required": []string{"access_token", "refresh_token", "access_expires_at"},
		},
		"AuthPayload": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"tokens": map[string]any{"$ref": "#/components/schemas/AuthTokens"},
				"user":   map[string]any{"$ref": "#/components/schemas/User"},
			},
			"required": []string{"tokens", "user"},
		},
		"RegisterRequest": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"email":     map[string]any{"type": "string", "format": "email"},
				"password":  map[string]any{"type": "string", "minLength": 8},
				"full_name": map[string]any{"type": "string"},
				"role":      map[string]any{"type": "string", "enum": []string{"owner", "worker", "moderator"}, "default": "worker"},
				"phone":     map[string]any{"type": "string"},
			},
			"required": []string{"email", "password", "full_name"},
		},
		"LoginRequest": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"email":    map[string]any{"type": "string", "format": "email"},
				"password": map[string]any{"type": "string"},
			},
			"required": []string{"email", "password"},
		},
		"RefreshRequest": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"refresh_token": map[string]any{"type": "string"},
			},
			"required": []string{"refresh_token"},
		},
		"LogoutRequest": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"refresh_token": map[string]any{"type": "string"},
			},
			"required": []string{"refresh_token"},
		},
		"UpdateProfileRequest": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"full_name":  map[string]any{"type": "string"},
				"phone":      map[string]any{"type": "string"},
				"avatar_url": map[string]any{"type": "string", "format": "uri"},
			},
		},
		"VerifyEmailRequest": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code": map[string]any{"type": "string", "example": "123456"},
			},
			"required": []string{"code"},
		},
		"ResendVerificationCodeResponse": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code_expires_at":     map[string]any{"type": "string", "format": "date-time", "nullable": true},
				"resend_available_at": map[string]any{"type": "string", "format": "date-time", "nullable": true},
			},
		},
	}
}

func paths() map[string]any {
	return map[string]any{
		"/health/live": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Health"},
				"summary":     "Liveness probe",
				"description": "Returns a lightweight process health signal.",
				"responses":   jsonResponses(200, "OK", "#/components/schemas/HealthStatus"),
			},
		},
		"/health/ready": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Health"},
				"summary":     "Readiness probe",
				"description": "Returns readiness for traffic acceptance.",
				"responses":   jsonResponses(200, "OK", "#/components/schemas/HealthStatus"),
			},
		},
		"/openapi.json": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Health"},
				"summary":     "OpenAPI document",
				"description": "Returns the generated OpenAPI 3.0 document for this gateway.",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "OpenAPI JSON",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"type": "object"},
							},
						},
					},
				},
			},
		},
		"/swagger": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Health"},
				"summary":     "Documentation UI",
				"description": "Serves the built-in interactive API documentation page for the gateway.",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "HTML documentation UI",
						"content": map[string]any{
							"text/html": map[string]any{
								"schema": map[string]any{"type": "string"},
							},
						},
					},
				},
			},
		},
		"/v1/auth/register": map[string]any{
			"post": op(opConfig{
				Tags:        []string{"Auth"},
				Summary:     "Register user",
				Description: "Creates a user account and returns JWT tokens. Unverified users are allowed to authenticate, but later blocked from non-whitelist endpoints.",
				RequestRef:  "#/components/schemas/RegisterRequest",
				ResponseRef: "#/components/schemas/AuthPayload",
				StatusCode:  201,
			}),
		},
		"/v1/auth/login": map[string]any{
			"post": op(opConfig{
				Tags:        []string{"Auth"},
				Summary:     "Login",
				Description: "Authenticates a user and returns fresh access and refresh tokens.",
				RequestRef:  "#/components/schemas/LoginRequest",
				ResponseRef: "#/components/schemas/AuthPayload",
				StatusCode:  200,
			}),
		},
		"/v1/auth/refresh": map[string]any{
			"post": op(opConfig{
				Tags:        []string{"Auth"},
				Summary:     "Refresh tokens",
				Description: "Exchanges a refresh token for a new auth payload.",
				RequestRef:  "#/components/schemas/RefreshRequest",
				ResponseRef: "#/components/schemas/AuthPayload",
				StatusCode:  200,
			}),
		},
		"/v1/auth/logout": map[string]any{
			"post": op(opConfig{
				Tags:          []string{"Auth"},
				Summary:       "Logout",
				Description:   "Revokes the refresh token. This endpoint is allowed for authenticated but unverified users.",
				RequestRef:    "#/components/schemas/LogoutRequest",
				StatusCode:    204,
				Auth:          true,
				AllowPending:  true,
				SuccessNoBody: true,
			}),
		},
		"/v1/auth/resend-code": map[string]any{
			"post": op(opConfig{
				Tags:         []string{"Auth"},
				Summary:      "Resend verification code",
				Description:  "Resends the current email verification code when cooldown allows. This endpoint is allowed for authenticated but unverified users.",
				ResponseRef:  "#/components/schemas/ResendVerificationCodeResponse",
				StatusCode:   200,
				Auth:         true,
				AllowPending: true,
			}),
		},
		"/v1/auth/verify-email": map[string]any{
			"post": op(opConfig{
				Tags:         []string{"Auth"},
				Summary:      "Verify email",
				Description:  "Verifies the current user's email using a code. This endpoint is allowed for authenticated but unverified users.",
				RequestRef:   "#/components/schemas/VerifyEmailRequest",
				ResponseRef:  "#/components/schemas/User",
				StatusCode:   200,
				Auth:         true,
				AllowPending: true,
			}),
		},
		"/v1/me": map[string]any{
			"get": op(opConfig{
				Tags:         []string{"Profile"},
				Summary:      "Get current user",
				Description:  "Returns the authenticated user's profile. Requires a verified email.",
				ResponseRef:  "#/components/schemas/User",
				StatusCode:   200,
				Auth:         true,
				VerifiedOnly: true,
			}),
			"patch": op(opConfig{
				Tags:         []string{"Profile"},
				Summary:      "Update current user",
				Description:  "Updates the authenticated user's profile fields that are present in the request body. Requires a verified email.",
				RequestRef:   "#/components/schemas/UpdateProfileRequest",
				ResponseRef:  "#/components/schemas/User",
				StatusCode:   200,
				Auth:         true,
				VerifiedOnly: true,
			}),
		},
		"/v1/pvz": map[string]any{
			"post": notImplementedOp("PVZ", "Create PVZ"),
			"get":  notImplementedOp("PVZ", "List owner PVZ"),
		},
		"/v1/pvz/{id}": map[string]any{
			"get":   notImplementedOp("PVZ", "Get PVZ by ID"),
			"patch": notImplementedOp("PVZ", "Patch PVZ"),
		},
		"/v1/pvz/{id}/submit-update": map[string]any{
			"post": notImplementedOp("PVZ", "Submit PVZ update"),
		},
		"/v1/moderation/pvz": map[string]any{
			"get": notImplementedOp("Moderation", "List PVZ moderation queue"),
		},
		"/v1/moderation/pvz/{id}/approve": map[string]any{
			"post": notImplementedOp("Moderation", "Approve PVZ"),
		},
		"/v1/moderation/pvz/{id}/reject": map[string]any{
			"post": notImplementedOp("Moderation", "Reject PVZ"),
		},
		"/v1/moderation/pvz/{id}/block": map[string]any{
			"post": notImplementedOp("Moderation", "Block PVZ"),
		},
		"/v1/vacancies": map[string]any{
			"post": notImplementedOp("Vacancies", "Create vacancy"),
			"get":  notImplementedOp("Vacancies", "List vacancies"),
		},
		"/v1/vacancies/{id}": map[string]any{
			"get": notImplementedOp("Vacancies", "Get vacancy by ID"),
		},
		"/v1/vacancies/{id}/responses": map[string]any{
			"post": notImplementedOp("Responses", "Create response"),
			"get":  notImplementedOp("Responses", "List responses for vacancy"),
		},
		"/v1/responses/{id}/withdraw": map[string]any{
			"post": notImplementedOp("Responses", "Withdraw response"),
		},
		"/v1/responses/{id}/accept": map[string]any{
			"post": notImplementedOp("Responses", "Accept response"),
		},
		"/v1/assignments/{id}/cancel": map[string]any{
			"post": notImplementedOp("Assignments", "Cancel assignment"),
		},
		"/v1/assignments/{id}/complete": map[string]any{
			"post": notImplementedOp("Assignments", "Complete assignment"),
		},
		"/v1/assignments/{id}/reviews": map[string]any{
			"post": notImplementedOp("Reviews", "Create review"),
		},
		"/v1/reviews": map[string]any{
			"get": notImplementedOp("Reviews", "List reviews"),
		},
		"/v1/reviews/{id}/replies": map[string]any{
			"post": notImplementedOp("Reviews", "Create review reply"),
		},
		"/v1/conversations": map[string]any{
			"get": notImplementedOp("Chat", "List conversations"),
		},
		"/v1/conversations/{id}/messages": map[string]any{
			"get":  notImplementedOp("Chat", "List conversation messages"),
			"post": notImplementedOp("Chat", "Send conversation message"),
		},
		"/v1/conversations/{id}/read": map[string]any{
			"post": notImplementedOp("Chat", "Mark conversation as read"),
		},
		"/ws/conversations/{id}": map[string]any{
			"get": map[string]any{
				"tags":        []string{"WebSocket"},
				"summary":     "Conversation WebSocket",
				"description": "WebSocket edge for real-time chat. The route exists but is currently a documented placeholder and returns `501 not_implemented`.",
				"security":    []map[string]any{{"bearerAuth": []string{}}},
				"responses":   errorResponses(501),
			},
		},
	}
}

type opConfig struct {
	Tags          []string
	Summary       string
	Description   string
	RequestRef    string
	ResponseRef   string
	StatusCode    int
	Auth          bool
	VerifiedOnly  bool
	AllowPending  bool
	SuccessNoBody bool
}

func op(cfg opConfig) map[string]any {
	item := map[string]any{
		"tags":        cfg.Tags,
		"summary":     cfg.Summary,
		"description": cfg.Description,
		"responses":   errorResponses(cfg.StatusCode),
	}

	if cfg.RequestRef != "" {
		item["requestBody"] = map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": cfg.RequestRef},
				},
			},
		}
	}

	if cfg.Auth {
		item["security"] = []map[string]any{{"bearerAuth": []string{}}}
	}

	if cfg.VerifiedOnly {
		item["x-berusmenu-verification"] = "verified_only"
	}
	if cfg.AllowPending {
		item["x-berusmenu-verification"] = "allowed_for_unverified"
	}

	responses := item["responses"].(map[string]any)
	code := statusCodeText(cfg.StatusCode)
	if cfg.SuccessNoBody {
		responses[code] = map[string]any{"description": "No Content"}
		return item
	}

	responses[code] = map[string]any{
		"description": "Success",
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": cfg.ResponseRef},
			},
		},
	}

	return item
}

func notImplementedOp(tag string, summary string) map[string]any {
	return map[string]any{
		"tags":                     []string{tag},
		"summary":                  summary,
		"description":              "Route is registered in the gateway, but the REST to gRPC mapping is not implemented yet.",
		"security":                 []map[string]any{{"bearerAuth": []string{}}},
		"x-berusmenu-verification": "verified_only",
		"responses":                errorResponses(501),
	}
}

func jsonResponses(code int, description string, schemaRef string) map[string]any {
	return map[string]any{
		statusCodeText(code): map[string]any{
			"description": description,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": schemaRef},
				},
			},
		},
	}
}

func errorResponses(successCode int) map[string]any {
	responses := map[string]any{
		"400": errorResponse("Bad Request"),
		"401": errorResponse("Unauthenticated"),
		"403": errorResponse("Forbidden"),
		"404": errorResponse("Not Found"),
		"409": errorResponse("Conflict"),
		"412": errorResponse("Failed Precondition"),
		"429": errorResponse("Too Many Requests"),
		"500": errorResponse("Internal Server Error"),
		"501": errorResponse("Not Implemented"),
		"503": errorResponse("Service Unavailable"),
		"504": errorResponse("Gateway Timeout"),
	}

	switch successCode {
	case 200, 201, 204:
		delete(responses, statusCodeText(successCode))
	}

	return responses
}

func errorResponse(description string) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
}

func statusCodeText(code int) string {
	switch code {
	case 200:
		return "200"
	case 201:
		return "201"
	case 204:
		return "204"
	case 501:
		return "501"
	default:
		return "200"
	}
}
