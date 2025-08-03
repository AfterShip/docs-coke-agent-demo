package code

// Common: basic errors. module code = 00
const (
	// ErrUnknown - 500: Internal server error.
	ErrUnknown uint32 = iota + SystemCode + InfraModelCode*1001

	// ErrBind - 400: Error occurred while binding the request body to the struct.
	ErrBind

	// ErrInvalidParams - 400: Invalid parameter.
	ErrInvalidParams

	// ErrValidation - 400: Validation failed.
	ErrValidation

	// ErrTokenInvalid - 401: Token invalid.
	ErrTokenInvalid

	// ErrPageNotFound - 404: Page not found.
	ErrPageNotFound

	// ErrTooManyRequests - 429: You have exceeded the API call rate limit.
	ErrTooManyRequests

	// ErrDatabase - 500: Database error.
	ErrDatabase

	// ErrEncrypt - 401: Error occurred while encrypting the user password.
	ErrEncrypt

	// ErrSignatureInvalid - 401: Signature is invalid.
	ErrSignatureInvalid

	// ErrExpired - 401: Token expired.
	ErrExpired

	// ErrInvalidAuthHeader - 401: Invalid authorization header.
	ErrInvalidAuthHeader

	// ErrMissingHeader - 401: The `Authorization` header was empty.
	ErrMissingHeader

	// ErrPasswordIncorrect - 401: Password was incorrect.
	ErrPasswordIncorrect

	// ErrPermissionDenied - 403: Permission denied.
	ErrPermissionDenied

	// ErrEncodingFailed - 500: Encoding failed due to an error with the data.
	ErrEncodingFailed

	// ErrDecodingFailed - 500: Decoding failed due to an error with the data.
	ErrDecodingFailed

	// ErrInvalidJSON - 500: Data is not valid JSON.
	ErrInvalidJSON

	// ErrEncodingJSON - 500: JSON data could not be encoded.
	ErrEncodingJSON

	// ErrDecodingJSON - 500: JSON data could not be decoded.
	ErrDecodingJSON

	// ErrInvalidYaml - 500: Data is not valid Yaml.
	ErrInvalidYaml

	// ErrEncodingYaml - 500: Yaml data could not be encoded.
	ErrEncodingYaml

	// ErrDecodingYaml - 500: Yaml data could not be decoded.
	ErrDecodingYaml

	// ErrLockWithFailed - 429: The resource that is being accessed is locked.Please wait for a moment and try again.
	ErrLockWithFailed
)
