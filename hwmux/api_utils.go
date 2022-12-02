package hwmux

import (
	"context"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"stash.silabs.com/iot_infra_sw/hwmux-client-golang"
)

// Get device, err and set error
func GetDevice(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	device *hwmux.DeviceSerializerPublic, httpRes *http.Response, err error) {
	device, httpRes, err = client.DevicesApi.DevicesRetrieve(context.Background(), id).Execute()
	if err != nil {
		diagnostics.AddError(
			"Unable to Read Device",
			err.Error(),
		)
	}
	return
}

// Get device, err and set error
func GetDeviceGroup(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	deviceGroup *hwmux.DeviceGroup, httpRes *http.Response, err error) {
	deviceGroup, httpRes, err = client.GroupsApi.GroupsRetrieve(context.Background(), id).Execute()
	if err != nil {
		diagnostics.AddError(
			"Unable to Read DeviceGroup",
			err.Error(),
		)
	}
	return
}

// Get Location by device id
func GetDeviceLocation(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	location *hwmux.Location, httpRes *http.Response, err error) {
	location, httpRes, err = client.DevicesApi.DevicesLocationRetrieve(context.Background(), strconv.Itoa(int(id))).Execute()
	if err != nil {
		diagnostics.AddError(
			"Unable to Read Device Location",
			err.Error(),
		)
	}
	return
}

// Get permission groups for a given deviceGroup
func GetPermissionGroupsForDeviceGroup(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	[]string, error) {
	var permissionGroups []string
	// TODO: Implement once API is available
	return permissionGroups, nil
}
