package hwmux

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Get string payload from io.ReadCloser
func BodyToString(body *io.ReadCloser) string {
	bodyStr := ""
	bytes, errB := io.ReadAll(*body)
	if errB == nil {
		bodyStr += "\n" + string(bytes)
	}
	return bodyStr
}

// Unmarshal metadata, set Diagnostics, return metadata and error
func UnmarshalMetadataSetError(data string, diagnostics *diag.Diagnostics, resourceName string) (*map[string]interface{}, error) {
	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(data), &metadata)
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Unable to decode %s metadata from json", resourceName),
			err.Error(),
		)
		return nil, err
	}
	return &metadata, nil
}

// Unmarshal metadata, set Diagnostics, set field and return error
func MarshalMetadataSetError(metadata map[string]interface{}, diagnostics *diag.Diagnostics, resourceName string, field *types.String) error {
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Unable to encode %s metadata to json", resourceName),
			err.Error(),
		)
		return err
	}
	*field = types.StringValue(string(metadataJson))
	return nil
}
