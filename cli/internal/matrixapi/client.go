package matrixapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HomeserverURL string
	AccessToken   string
	HTTPClient    *http.Client
}

func NewClient(homeserverURL, accessToken string) *Client {
	return &Client{
		HomeserverURL: homeserverURL,
		AccessToken:   accessToken,
		HTTPClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Login(user, password string) (string, error) {
	body := map[string]any{
		"type": "m.login.password",
		"identifier": map[string]string{
			"type": "m.id.user",
			"user": user,
		},
		"password": password,
	}

	var resp struct {
		AccessToken string `json:"access_token"`
		UserID      string `json:"user_id"`
	}

	if err := c.post("/_matrix/client/v3/login", body, &resp); err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}

	c.AccessToken = resp.AccessToken
	return resp.AccessToken, nil
}

func (c *Client) WhoAmI() (string, error) {
	var resp struct {
		UserID string `json:"user_id"`
	}
	if err := c.get("/_matrix/client/v3/account/whoami", &resp); err != nil {
		return "", err
	}
	return resp.UserID, nil
}

func (c *Client) CreateRoom(name, topic string, invite []string, isDirect bool) (string, error) {
	body := map[string]any{
		"name":       name,
		"topic":      topic,
		"visibility": "private",
		"preset":     "private_chat",
		"is_direct":  isDirect,
	}
	if len(invite) > 0 {
		body["invite"] = invite
	}

	var resp struct {
		RoomID string `json:"room_id"`
	}
	if err := c.post("/_matrix/client/v3/createRoom", body, &resp); err != nil {
		return "", fmt.Errorf("create room failed: %w", err)
	}
	return resp.RoomID, nil
}

func (c *Client) JoinRoom(roomIDOrAlias string) (string, error) {
	var resp struct {
		RoomID string `json:"room_id"`
	}
	if err := c.post("/_matrix/client/v3/join/"+roomIDOrAlias, map[string]any{}, &resp); err != nil {
		return "", fmt.Errorf("join room failed: %w", err)
	}
	return resp.RoomID, nil
}

func (c *Client) LeaveRoom(roomID string) error {
	return c.post("/_matrix/client/v3/rooms/"+roomID+"/leave", map[string]any{}, nil)
}

func (c *Client) InviteUser(roomID, userID string) error {
	body := map[string]string{"user_id": userID}
	return c.post("/_matrix/client/v3/rooms/"+roomID+"/invite", body, nil)
}

func (c *Client) SendMessage(roomID, message string) (string, error) {
	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	body := map[string]string{
		"msgtype": "m.text",
		"body":    message,
	}

	var resp struct {
		EventID string `json:"event_id"`
	}
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/m.room.message/%s", roomID, txnID)
	if err := c.put(path, body, &resp); err != nil {
		return "", fmt.Errorf("send message failed: %w", err)
	}
	return resp.EventID, nil
}

func (c *Client) JoinedRooms() ([]string, error) {
	var resp struct {
		JoinedRooms []string `json:"joined_rooms"`
	}
	if err := c.get("/_matrix/client/v3/joined_rooms", &resp); err != nil {
		return nil, err
	}
	return resp.JoinedRooms, nil
}

// RoomMessages fetches message history from a room.
type RoomMessage struct {
	Type     string         `json:"type"`
	Sender   string         `json:"sender"`
	EventID  string         `json:"event_id"`
	Content  map[string]any `json:"content"`
	OriginTS int64          `json:"origin_server_ts"`
}

type MessagesResponse struct {
	Chunk []RoomMessage `json:"chunk"`
	End   string        `json:"end"`
	Start string        `json:"start"`
}

func (c *Client) RoomMessages(roomID, from, dir string, limit int) (*MessagesResponse, error) {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/messages?dir=%s&limit=%d", roomID, dir, limit)
	if from != "" {
		path += "&from=" + from
	}
	var resp MessagesResponse
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Sync performs a long-poll sync. Returns the response and the next_batch token.
type SyncResponse struct {
	NextBatch string                       `json:"next_batch"`
	Rooms     struct {
		Join map[string]SyncJoinedRoom `json:"join"`
	} `json:"rooms"`
}

type SyncJoinedRoom struct {
	Timeline struct {
		Events []RoomMessage `json:"events"`
	} `json:"timeline"`
}

func (c *Client) Sync(since string, timeout int) (*SyncResponse, error) {
	path := fmt.Sprintf("/_matrix/client/v3/sync?timeout=%d", timeout)
	if since != "" {
		path += "&since=" + since
	}
	// Use a longer HTTP timeout than the sync timeout
	old := c.HTTPClient.Timeout
	c.HTTPClient.Timeout = time.Duration(timeout+10000) * time.Millisecond
	defer func() { c.HTTPClient.Timeout = old }()

	var resp SyncResponse
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetRoomAlias sets a canonical alias for a room.
func (c *Client) SetRoomAlias(alias, roomID string) error {
	path := fmt.Sprintf("/_matrix/client/v3/directory/room/%s", alias)
	body := map[string]string{"room_id": roomID}
	return c.put(path, body, nil)
}

// DeleteRoomAlias removes a room alias.
func (c *Client) DeleteRoomAlias(alias string) error {
	path := fmt.Sprintf("/_matrix/client/v3/directory/room/%s", alias)
	req, err := http.NewRequest("DELETE", c.HomeserverURL+path, nil)
	if err != nil {
		return err
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}
	return c.doRequest(req, nil)
}

// ResolveAlias resolves a room alias to a room ID.
func (c *Client) ResolveAlias(alias string) (string, error) {
	path := fmt.Sprintf("/_matrix/client/v3/directory/room/%s", alias)
	var resp struct {
		RoomID string `json:"room_id"`
	}
	if err := c.get(path, &resp); err != nil {
		return "", err
	}
	return resp.RoomID, nil
}

// SendFile sends a file message to a room (after uploading via UploadMedia).
func (c *Client) SendFile(roomID, mxcURL, filename, mimetype string, size int64) (string, error) {
	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	body := map[string]any{
		"msgtype": "m.file",
		"body":    filename,
		"url":     mxcURL,
		"info": map[string]any{
			"mimetype": mimetype,
			"size":     size,
		},
	}
	var resp struct {
		EventID string `json:"event_id"`
	}
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/m.room.message/%s", roomID, txnID)
	if err := c.put(path, body, &resp); err != nil {
		return "", fmt.Errorf("send file failed: %w", err)
	}
	return resp.EventID, nil
}

// SendImage sends an image message to a room.
func (c *Client) SendImage(roomID, mxcURL, filename, mimetype string, size int64) (string, error) {
	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	body := map[string]any{
		"msgtype": "m.image",
		"body":    filename,
		"url":     mxcURL,
		"info": map[string]any{
			"mimetype": mimetype,
			"size":     size,
		},
	}
	var resp struct {
		EventID string `json:"event_id"`
	}
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/m.room.message/%s", roomID, txnID)
	if err := c.put(path, body, &resp); err != nil {
		return "", fmt.Errorf("send image failed: %w", err)
	}
	return resp.EventID, nil
}

// UploadMedia uploads a file and returns the mxc:// URL.
func (c *Client) UploadMedia(data []byte, filename, contentType string) (string, error) {
	path := fmt.Sprintf("/_matrix/media/v3/upload?filename=%s", filename)
	req, err := http.NewRequest("POST", c.HomeserverURL+path, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}
	var resp struct {
		ContentURI string `json:"content_uri"`
	}
	if err := c.doRequest(req, &resp); err != nil {
		return "", err
	}
	return resp.ContentURI, nil
}

func (c *Client) ServerVersion() (string, error) {
	var resp struct {
		Server struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"server"`
	}
	if err := c.getNoAuth("/_matrix/federation/v1/version", &resp); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", resp.Server.Name, resp.Server.Version), nil
}

func (c *Client) get(path string, result any) error {
	req, err := http.NewRequest("GET", c.HomeserverURL+path, nil)
	if err != nil {
		return err
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}
	return c.doRequest(req, result)
}

func (c *Client) getNoAuth(path string, result any) error {
	req, err := http.NewRequest("GET", c.HomeserverURL+path, nil)
	if err != nil {
		return err
	}
	return c.doRequest(req, result)
}

func (c *Client) post(path string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.HomeserverURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}
	return c.doRequest(req, result)
}

func (c *Client) put(path string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", c.HomeserverURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}
	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			ErrCode string `json:"errcode"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s: %s", errResp.ErrCode, errResp.Error)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}
	return nil
}
