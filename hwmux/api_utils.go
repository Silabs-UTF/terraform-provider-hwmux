package hwmux

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Silabs-UTF/hwmux-client-golang"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Get device, err and set error
func GetDevice(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	device *hwmux.DeviceSerializerPublic, httpRes *http.Response, err error) {
	device, httpRes, err = client.DevicesApi.DevicesRetrieve(context.Background(), id).Execute()
	handleError(httpRes, err, diagnostics, "Device")
	return
}

// Get deviceGroup, err and set error
func GetDeviceGroup(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	deviceGroup *hwmux.DeviceGroup, httpRes *http.Response, err error) {
	deviceGroup, httpRes, err = client.GroupsApi.GroupsRetrieve(context.Background(), id).Execute()
	handleError(httpRes, err, diagnostics, "Device Group")
	return
}

// Get label, err and set error
func GetLabel(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	label *hwmux.Label, httpRes *http.Response, err error) {
	label, httpRes, err = client.LabelsApi.LabelsRetrieve(context.Background(), id).Execute()
	handleError(httpRes, err, diagnostics, "Label")
	return
}

// Get part, err and set error
func GetPart(client *hwmux.APIClient, diagnostics *diag.Diagnostics, part_no string) (
	part *hwmux.Part, httpRes *http.Response, err error) {
	part, httpRes, err = client.PartsApi.PartsRetrieve(context.Background(), part_no).Execute()
	handleError(httpRes, err, diagnostics, "Part")
	return
}

// Get room, err and set error
func GetRoom(client *hwmux.APIClient, diagnostics *diag.Diagnostics, name string) (
	room *hwmux.Room, httpRes *http.Response, err error) {
	room, httpRes, err = client.RoomsApi.RoomsRetrieve(context.Background(), name).Execute()
	handleError(httpRes, err, diagnostics, "Room")
	return
}

// Get permission group, err and set error
func GetPermissionGroup(client *hwmux.APIClient, diagnostics *diag.Diagnostics, name string) (
	permissionGroup *hwmux.PermissionGroup, httpRes *http.Response, err error) {
	permissionGroup, httpRes, err = client.PermissionsApi.PermissionsGroupsRetrieve(context.Background(), name).Execute()
	handleError(httpRes, err, diagnostics, "Permission Group")
	return
}

// Get user, err and set error
func GetUser(client *hwmux.APIClient, diagnostics *diag.Diagnostics, username string) (
	user *hwmux.LoggedInUser, httpRes *http.Response, err error) {
	user, httpRes, err = client.UserApi.UserRetrieve(context.Background(), username).Execute()
	handleError(httpRes, err, diagnostics, "User")
	return
}

// Get Location by device id
func GetDeviceLocation(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	location *hwmux.Location, httpRes *http.Response, err error) {
	location, httpRes, err = client.DevicesApi.DevicesLocationRetrieve(context.Background(), strconv.Itoa(int(id))).Execute()
	handleError(httpRes, err, diagnostics, "Device Location")
	return
}

// Get permission groups for a given deviceGroup
func GetPermissionGroupsForDeviceGroup(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	[]string, error) {
	objectPerms, httpRes, err := client.GroupsApi.GroupsPermissionsRetrieve(context.Background(), id).Execute()
	handleError(httpRes, err, diagnostics, "Permissions for Device Group")
	return objectPermsToUGList(objectPerms), nil
}

// Get permission groups for a given device
func GetPermissionGroupsForDevice(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	[]string, error) {
	objectPerms, httpRes, err := client.DevicesApi.DevicesPermissionsRetrieve(context.Background(), id).Execute()
	handleError(httpRes, err, diagnostics, "Permissions for Device")
	return objectPermsToUGList(objectPerms), nil
}


// Get permission groups for a given Label
func GetPermissionGroupsForLabel(client *hwmux.APIClient, diagnostics *diag.Diagnostics, id int32) (
	[]string, error) {
	objectPerms, httpRes, err := client.LabelsApi.LabelsPermissionsRetrieve(context.Background(), id).Execute()
	handleError(httpRes, err, diagnostics, "Permissions for Label")
	return objectPermsToUGList(objectPerms), nil
}


// Returns all user group names from the given object permissions object
func objectPermsToUGList(objectPerms *hwmux.ObjectPermissions) []string {
	permissionGroups := make([]string, len(objectPerms.GetUserGroups()))
	i := 0
	for key := range objectPerms.GetUserGroups() {
		permissionGroups[i] = key
		i++
	}
	return permissionGroups
}


// factored out error handling code for API retrieve calls
func handleError(httpRes *http.Response, err error, diagnostics *diag.Diagnostics, name string) {
	if err != nil {
		errorStr := err.Error()
		if httpRes != nil {
			errorStr += "\nHwmux response body:" + BodyToString(&httpRes.Body)
		}
		diagnostics.AddError(
			"Unable to Read " + name,
			errorStr,
		)
	}
}
