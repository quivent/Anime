package matrixapi

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

type AdminClient struct {
	*Client
	Domain string
}

func NewAdminClient(homeserverURL, adminToken, domain string) *AdminClient {
	return &AdminClient{
		Client: NewClient(homeserverURL, adminToken),
		Domain: domain,
	}
}

type UserInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayname"`
	Admin       int    `json:"admin"`
	Deactivated int    `json:"deactivated"`
	CreationTS  int64  `json:"creation_ts"`
	AvatarURL   string `json:"avatar_url"`
}

type UserListResponse struct {
	Users     []UserInfo `json:"users"`
	NextToken string     `json:"next_token"`
	Total     int        `json:"total"`
}

func (a *AdminClient) ListUsers(from int, limit int) (*UserListResponse, error) {
	path := fmt.Sprintf("/_synapse/admin/v2/users?from=%d&limit=%d", from, limit)
	var resp UserListResponse
	if err := a.get(path, &resp); err != nil {
		return nil, fmt.Errorf("list users failed: %w", err)
	}
	return &resp, nil
}

func (a *AdminClient) CreateUser(localpart, password, displayName string, admin bool) error {
	path := fmt.Sprintf("/_synapse/admin/v2/users/@%s:%s", localpart, a.Domain)
	body := map[string]any{
		"password":    password,
		"displayname": displayName,
		"admin":       admin,
		"deactivated": false,
	}
	return a.put(path, body, nil)
}

func (a *AdminClient) DeactivateUser(userID string) error {
	path := fmt.Sprintf("/_synapse/admin/v1/deactivate/%s", userID)
	body := map[string]any{"erase": false}
	return a.post(path, body, nil)
}

func (a *AdminClient) SetAdmin(userID string, admin bool) error {
	path := fmt.Sprintf("/_synapse/admin/v2/users/%s", userID)
	body := map[string]any{"admin": admin}
	return a.put(path, body, nil)
}

func (a *AdminClient) ResetPassword(userID, newPassword string) error {
	path := fmt.Sprintf("/_synapse/admin/v1/reset_password/%s", userID)
	body := map[string]any{
		"new_password":   newPassword,
		"logout_devices": true,
	}
	return a.post(path, body, nil)
}

func (a *AdminClient) GetUserInfo(userID string) (*UserInfo, error) {
	path := fmt.Sprintf("/_synapse/admin/v2/users/%s", userID)
	var resp UserInfo
	if err := a.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *AdminClient) RegisterWithSharedSecret(user, password, sharedSecret string, admin bool) error {
	var nonceResp struct {
		Nonce string `json:"nonce"`
	}
	if err := a.getNoAuth("/_synapse/admin/v1/register", &nonceResp); err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	adminStr := "notadmin"
	if admin {
		adminStr = "admin"
	}
	mac := hmac.New(sha1.New, []byte(sharedSecret))
	mac.Write([]byte(nonceResp.Nonce + "\x00" + user + "\x00" + password + "\x00" + adminStr))

	body := map[string]any{
		"nonce":    nonceResp.Nonce,
		"username": user,
		"password": password,
		"admin":    admin,
		"mac":      hex.EncodeToString(mac.Sum(nil)),
	}
	return a.post("/_synapse/admin/v1/register", body, nil)
}

type RoomInfo struct {
	RoomID         string `json:"room_id"`
	Name           string `json:"name"`
	Topic          string `json:"topic"`
	NumMembers     int    `json:"num_joined_members"`
	Creator        string `json:"creator"`
	CanonicalAlias string `json:"canonical_alias"`
}

type RoomListResponse struct {
	Rooms     []RoomInfo `json:"rooms"`
	NextBatch string     `json:"next_batch"`
	Total     int        `json:"total_rooms"`
}

func (a *AdminClient) ListRooms(from int, limit int) (*RoomListResponse, error) {
	path := fmt.Sprintf("/_synapse/admin/v1/rooms?from=%d&limit=%d&dir=f", from, limit)
	var resp RoomListResponse
	if err := a.get(path, &resp); err != nil {
		return nil, fmt.Errorf("list rooms failed: %w", err)
	}
	return &resp, nil
}
