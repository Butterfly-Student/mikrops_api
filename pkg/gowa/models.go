package gowa

import "time"

// GenericResponse is the base response structure for all Gowa API responses.
type GenericResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Results interface{} `json:"results"`
}

// ErrorResponse represents an error response from the Gowa API.
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Results interface{} `json:"results"`
}

// LoginResponse is the response for the login endpoint.
type LoginResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Results LoginResult `json:"results"`
}

// LoginResult contains the QR code information for login.
type LoginResult struct {
	QRDuration int    `json:"qr_duration"`
	QRLink     string `json:"qr_link"`
}

// LoginWithCodeResponse is the response for the login with pairing code endpoint.
type LoginWithCodeResponse struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Results LoginWithCodeResult `json:"results"`
}

// LoginWithCodeResult contains the pairing code for login.
type LoginWithCodeResult struct {
	PairCode string `json:"pair_code"`
}

// AppStatusResponse is the response for the app status endpoint.
type AppStatusResponse struct {
	Status  int             `json:"status"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Results AppStatusResult `json:"results"`
}

// AppStatusResult contains the connection status information.
type AppStatusResult struct {
	IsConnected bool   `json:"is_connected"`
	IsLoggedIn  bool   `json:"is_logged_in"`
	DeviceID    string `json:"device_id"`
}

// DeviceInfo represents a device in the Gowa system.
type DeviceInfo struct {
	ID          string    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	DisplayName string    `json:"display_name"`
	State       string    `json:"state"` // disconnected, connected, logged_in
	JID         string    `json:"jid"`
	CreatedAt   time.Time `json:"created_at"`
}

// DeviceListResponse is the response for listing all devices.
type DeviceListResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Status  int          `json:"status"`
	Results []DeviceInfo `json:"results"`
}

// DeviceAddResponse is the response for adding a new device.
type DeviceAddResponse struct {
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Status  int        `json:"status"`
	Results DeviceInfo `json:"results"`
}

// DeviceInfoResponse is the response for getting device info.
type DeviceInfoResponse struct {
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Status  int        `json:"status"`
	Results DeviceInfo `json:"results"`
}

// DeviceStatusResponse is the response for getting device status.
type DeviceStatusResponse struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Status  int                `json:"status"`
	Results DeviceStatusResult `json:"results"`
}

// DeviceStatusResult contains the device connection status.
type DeviceStatusResult struct {
	DeviceID    string `json:"device_id"`
	IsConnected bool   `json:"is_connected"`
	IsLoggedIn  bool   `json:"is_logged_in"`
}

// SendMessageRequest is the request body for sending a text message.
type SendMessageRequest struct {
	Phone          string   `json:"phone"`
	Message        string   `json:"message"`
	ReplyMessageID string   `json:"reply_message_id,omitempty"`
	IsForwarded    bool     `json:"is_forwarded,omitempty"`
	Duration       int      `json:"duration,omitempty"`
	Mentions       []string `json:"mentions,omitempty"`
}

// SendResponse is the response for sending a message.
type SendResponse struct {
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Results SendResult `json:"results"`
}

// SendResult contains the message ID and status after sending.
type SendResult struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

// Group represents a WhatsApp group.
type Group struct {
	JID                           string        `json:"JID"`
	OwnerJID                      string        `json:"OwnerJID"`
	Name                          string        `json:"Name"`
	NameSetAt                     time.Time     `json:"NameSetAt"`
	NameSetBy                     string        `json:"NameSetBy"`
	Topic                         string        `json:"Topic"`
	TopicID                       string        `json:"TopicID"`
	TopicSetAt                    time.Time     `json:"TopicSetAt"`
	TopicSetBy                    string        `json:"TopicSetBy"`
	TopicDeleted                  bool          `json:"TopicDeleted"`
	IsLocked                      bool          `json:"IsLocked"`
	IsAnnounce                    bool          `json:"IsAnnounce"`
	AnnounceVersionID             string        `json:"AnnounceVersionID"`
	IsEphemeral                   bool          `json:"IsEphemeral"`
	DisappearingTimer             int           `json:"DisappearingTimer"`
	IsIncognito                   bool          `json:"IsIncognito"`
	IsParent                      bool          `json:"IsParent"`
	DefaultMembershipApprovalMode string        `json:"DefaultMembershipApprovalMode"`
	LinkedParentJID               string        `json:"LinkedParentJID"`
	IsDefaultSubGroup             bool          `json:"IsDefaultSubGroup"`
	IsJoinApprovalRequired        bool          `json:"IsJoinApprovalRequired"`
	GroupCreated                  time.Time     `json:"GroupCreated"`
	ParticipantVersionID          string        `json:"ParticipantVersionID"`
	Participants                  []Participant `json:"Participants"`
	MemberAddMode                 string        `json:"MemberAddMode"`
}

// Participant represents a participant in a WhatsApp group.
type Participant struct {
	JID          string `json:"JID"`
	LID          string `json:"LID"`
	IsAdmin      bool   `json:"IsAdmin"`
	IsSuperAdmin bool   `json:"IsSuperAdmin"`
	DisplayName  string `json:"DisplayName"`
	Error        int    `json:"Error"`
}

// UserGroupResponse is the response for getting user's groups.
type UserGroupResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Results UserGroupResult `json:"results"`
}

// UserGroupResult contains the list of groups.
type UserGroupResult struct {
	Data []Group `json:"data"`
}

// GroupResponse is the response for group operations.
type GroupResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Results GroupResult `json:"results"`
}

// GroupResult contains the list of groups.
type GroupResult struct {
	Data []Group `json:"data"`
}

// UserInfoResponse is the response for getting user info.
type UserInfoResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Results UserInfoResult `json:"results"`
}

// UserInfoResult contains user information.
type UserInfoResult struct {
	Devices []UserDevice `json:"devices"`
}

// UserDevice represents a user's device.
type UserDevice struct {
	User   string `json:"User"`
	Agent  int    `json:"Agent"`
	Device string `json:"Device"`
	Server string `json:"Server"`
	AD     bool   `json:"AD"`
}

// UserCheckResponse is the response for checking if a user is on WhatsApp.
type UserCheckResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Results UserCheckResult `json:"results"`
}

// UserCheckResult contains the check result.
type UserCheckResult struct {
	IsOnWhatsApp bool `json:"is_on_whatsapp"`
}

// GroupInfoResponse is the response for getting group info.
type GroupInfoResponse struct {
	Status  int         `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Results interface{} `json:"results"`
}

// GetGroupInviteLinkResponse is the response for getting a group invite link.
type GetGroupInviteLinkResponse struct {
	Code    string                `json:"code"`
	Message string                `json:"message"`
	Results GroupInviteLinkResult `json:"results"`
}

// GroupInviteLinkResult contains the invite link.
type GroupInviteLinkResult struct {
	InviteLink string `json:"invite_link"`
	GroupID    string `json:"group_id"`
}

// AddDeviceRequest is the request body for adding a new device.
type AddDeviceRequest struct {
	DeviceID string `json:"device_id,omitempty"`
}
