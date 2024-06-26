// Code generated by go-swagger; DO NOT EDIT.

// Copyright Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package silence

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/gh-liu/alertmanager/api/v2/models"
)

// GetSilencesReader is a Reader for the GetSilences structure.
type GetSilencesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetSilencesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetSilencesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewGetSilencesInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetSilencesOK creates a GetSilencesOK with default headers values
func NewGetSilencesOK() *GetSilencesOK {
	return &GetSilencesOK{}
}

/*GetSilencesOK handles this case with default header values.

Get silences response
*/
type GetSilencesOK struct {
	Payload models.GettableSilences
}

func (o *GetSilencesOK) Error() string {
	return fmt.Sprintf("[GET /silences][%d] getSilencesOK  %+v", 200, o.Payload)
}

func (o *GetSilencesOK) GetPayload() models.GettableSilences {
	return o.Payload
}

func (o *GetSilencesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSilencesInternalServerError creates a GetSilencesInternalServerError with default headers values
func NewGetSilencesInternalServerError() *GetSilencesInternalServerError {
	return &GetSilencesInternalServerError{}
}

/*GetSilencesInternalServerError handles this case with default header values.

Internal server error
*/
type GetSilencesInternalServerError struct {
	Payload string
}

func (o *GetSilencesInternalServerError) Error() string {
	return fmt.Sprintf("[GET /silences][%d] getSilencesInternalServerError  %+v", 500, o.Payload)
}

func (o *GetSilencesInternalServerError) GetPayload() string {
	return o.Payload
}

func (o *GetSilencesInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
