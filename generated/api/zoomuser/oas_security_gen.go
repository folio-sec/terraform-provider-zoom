// Code generated by ogen, DO NOT EDIT.

package zoomuser

import (
	"context"
	"net/http"

	"github.com/go-faster/errors"
)

// SecuritySource is provider of security values (tokens, passwords, etc.).
type SecuritySource interface {
	// OpenapiAuthorization provides openapi_authorization security value.
	OpenapiAuthorization(ctx context.Context, operationName OperationName) (OpenapiAuthorization, error)
	// OpenapiOAuth provides openapi_oauth security value.
	OpenapiOAuth(ctx context.Context, operationName OperationName) (OpenapiOAuth, error)
}

func (s *Client) securityOpenapiAuthorization(ctx context.Context, operationName OperationName, req *http.Request) error {
	t, err := s.sec.OpenapiAuthorization(ctx, operationName)
	if err != nil {
		return errors.Wrap(err, "security source \"OpenapiAuthorization\"")
	}
	req.Header.Set("Authorization", t.APIKey)
	return nil
}
func (s *Client) securityOpenapiOAuth(ctx context.Context, operationName OperationName, req *http.Request) error {
	t, err := s.sec.OpenapiOAuth(ctx, operationName)
	if err != nil {
		return errors.Wrap(err, "security source \"OpenapiOAuth\"")
	}
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return nil
}
