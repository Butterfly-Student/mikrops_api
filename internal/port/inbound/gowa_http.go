package inbound_port

import "github.com/gin-gonic/gin"

type GowaHttpPort interface {
	// Device management
	ListDevices(c *gin.Context)
	AddDevice(c *gin.Context)
	GetDevice(c *gin.Context)
	RemoveDevice(c *gin.Context)
	LoginDevice(c *gin.Context)
	LoginDeviceWithCode(c *gin.Context)
	LogoutDevice(c *gin.Context)
	ReconnectDevice(c *gin.Context)
	GetDeviceStatus(c *gin.Context)

	// App / session management (default or specified device via ?device_id=)
	AppLogin(c *gin.Context)
	AppLoginWithCode(c *gin.Context)
	AppLogout(c *gin.Context)
	AppReconnect(c *gin.Context)
	AppStatus(c *gin.Context)

	// Group management
	GetMyGroups(c *gin.Context)
	FindGroupByName(c *gin.Context)
	GetGroupInfo(c *gin.Context)
	GetGroupInviteLink(c *gin.Context)

	// User information
	CheckUser(c *gin.Context)
	GetUserInfo(c *gin.Context)
	GetMyContacts(c *gin.Context)

	// Send messages
	SendMessage(c *gin.Context)
	SendImageFromURL(c *gin.Context)
	SendFileFromURL(c *gin.Context)
	SendVideoFromURL(c *gin.Context)
}
