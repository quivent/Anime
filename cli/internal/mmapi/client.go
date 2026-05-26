package mmapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

// ─── Types ───────────────────────────────────────────────────────────────────

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Roles     string `json:"roles"`
	CreateAt  int64  `json:"create_at"`
	DeleteAt  int64  `json:"delete_at"`
}

func (u User) IsAdmin() bool   { return strings.Contains(u.Roles, "system_admin") }
func (u User) IsDeleted() bool { return u.DeleteAt > 0 }

type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"` // "O" open, "I" invite
}

type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"` // "O" public, "P" private, "D" direct
	TeamID      string `json:"team_id"`
	Purpose     string `json:"purpose"`
	Header      string `json:"header"`
	MemberCount int    `json:"member_count"`
	CreateAt    int64  `json:"create_at"`
}

type Post struct {
	ID        string   `json:"id"`
	ChannelID string   `json:"channel_id"`
	UserID    string   `json:"user_id"`
	Message   string   `json:"message"`
	Type      string   `json:"type"`
	CreateAt  int64    `json:"create_at"`
	FileIDs   []string `json:"file_ids,omitempty"`
}

type PostList struct {
	Order map[string]int `json:"order_map,omitempty"`
	Posts map[string]Post `json:"posts"`
	// Order as returned by API: newest first
	OrderArr []string `json:"order"`
}

type FileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mmError struct {
	Message   string `json:"message"`
	DetailedError string `json:"detailed_error"`
	StatusCode int    `json:"status_code"`
	ID        string `json:"id"`
}

// ─── Client ──────────────────────────────────────────────────────────────────

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.BaseURL+"/api/v4"+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

func (c *Client) decode(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var mmErr mmError
		_ = json.NewDecoder(resp.Body).Decode(&mmErr)
		if mmErr.Message != "" {
			return fmt.Errorf("%s (status %d)", mmErr.Message, resp.StatusCode)
		}
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) get(path string, out interface{}) error {
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return err
	}
	return c.decode(resp, out)
}

func (c *Client) post(path string, body, out interface{}) error {
	resp, err := c.do("POST", path, body)
	if err != nil {
		return err
	}
	return c.decode(resp, out)
}

func (c *Client) put(path string, body interface{}) error {
	resp, err := c.do("PUT", path, body)
	if err != nil {
		return err
	}
	return c.decode(resp, nil)
}

func (c *Client) delete(path string) error {
	resp, err := c.do("DELETE", path, nil)
	if err != nil {
		return err
	}
	return c.decode(resp, nil)
}

// ─── Auth ────────────────────────────────────────────────────────────────────

func (c *Client) Login(loginID, password string) (string, error) {
	body := map[string]string{"login_id": loginID, "password": password}
	resp, err := c.do("POST", "/users/login", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var mmErr mmError
		_ = json.NewDecoder(resp.Body).Decode(&mmErr)
		if mmErr.Message != "" {
			return "", fmt.Errorf("login failed: %s", mmErr.Message)
		}
		return "", fmt.Errorf("login failed: HTTP %d", resp.StatusCode)
	}
	token := resp.Header.Get("Token")
	if token == "" {
		return "", fmt.Errorf("no token in login response")
	}
	return token, nil
}

func (c *Client) GetMe() (*User, error) {
	var u User
	return &u, c.get("/users/me", &u)
}

func (c *Client) Ping() error {
	resp, err := c.do("GET", "/system/ping", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// ─── Users ───────────────────────────────────────────────────────────────────

func (c *Client) CreateUser(username, email, password string) (*User, error) {
	body := map[string]string{"username": username, "email": email, "password": password}
	var u User
	return &u, c.post("/users", body, &u)
}

func (c *Client) ListUsers(page, perPage int) ([]User, error) {
	var users []User
	return users, c.get(fmt.Sprintf("/users?page=%d&per_page=%d", page, perPage), &users)
}

func (c *Client) SearchUsers(term string, limit int) ([]User, error) {
	body := map[string]interface{}{"term": term, "limit": limit}
	var users []User
	return users, c.post("/users/search", body, &users)
}

func (c *Client) GetUser(userID string) (*User, error) {
	var u User
	return &u, c.get("/users/"+userID, &u)
}

func (c *Client) GetUserByUsername(username string) (*User, error) {
	var u User
	return &u, c.get("/users/username/"+username, &u)
}

func (c *Client) DeactivateUser(userID string) error {
	return c.delete("/users/" + userID)
}

func (c *Client) SetAdmin(userID string, isAdmin bool) error {
	roles := "system_user"
	if isAdmin {
		roles = "system_admin system_user"
	}
	return c.put("/users/"+userID+"/roles", map[string]string{"roles": roles})
}

func (c *Client) ResetPassword(userID, newPassword string) error {
	return c.put("/users/"+userID+"/password", map[string]string{
		"new_password": newPassword,
	})
}

// CreatePersonalToken creates a personal access token for a user.
func (c *Client) CreatePersonalToken(userID, description string) (string, error) {
	var result struct {
		Token string `json:"token"`
	}
	err := c.post("/users/"+userID+"/tokens", map[string]string{"description": description}, &result)
	return result.Token, err
}

// ─── Teams ───────────────────────────────────────────────────────────────────

func (c *Client) GetTeams(page, perPage int) ([]Team, error) {
	var teams []Team
	return teams, c.get(fmt.Sprintf("/teams?page=%d&per_page=%d", page, perPage), &teams)
}

func (c *Client) CreateTeam(name, displayName string) (*Team, error) {
	body := map[string]string{"name": name, "display_name": displayName, "type": "O"}
	var t Team
	return &t, c.post("/teams", body, &t)
}

func (c *Client) GetTeamByName(name string) (*Team, error) {
	var t Team
	return &t, c.get("/teams/name/"+name, &t)
}

func (c *Client) AddTeamMember(teamID, userID string) error {
	return c.post("/teams/"+teamID+"/members", map[string]string{
		"team_id": teamID, "user_id": userID,
	}, nil)
}

// EnsureTeam finds or creates a team, returning its ID.
func (c *Client) EnsureTeam(name, displayName string) (string, error) {
	t, err := c.GetTeamByName(name)
	if err == nil {
		return t.ID, nil
	}
	t, err = c.CreateTeam(name, displayName)
	if err != nil {
		return "", err
	}
	return t.ID, nil
}

// ─── Channels ────────────────────────────────────────────────────────────────

func (c *Client) GetTeamChannels(teamID string, page, perPage int) ([]Channel, error) {
	var channels []Channel
	return channels, c.get(fmt.Sprintf("/channels?team_id=%s&page=%d&per_page=%d", teamID, page, perPage), &channels)
}

func (c *Client) CreateChannel(teamID, name, displayName, channelType, purpose string) (*Channel, error) {
	body := map[string]string{
		"team_id":      teamID,
		"name":         name,
		"display_name": displayName,
		"type":         channelType,
		"purpose":      purpose,
	}
	var ch Channel
	return &ch, c.post("/channels", body, &ch)
}

func (c *Client) GetChannel(channelID string) (*Channel, error) {
	var ch Channel
	return &ch, c.get("/channels/"+channelID, &ch)
}

func (c *Client) GetChannelByName(teamID, name string) (*Channel, error) {
	var ch Channel
	return &ch, c.get("/teams/"+teamID+"/channels/name/"+name, &ch)
}

func (c *Client) AddChannelMember(channelID, userID string) error {
	return c.post("/channels/"+channelID+"/members", map[string]string{"user_id": userID}, nil)
}

func (c *Client) RemoveChannelMember(channelID, userID string) error {
	return c.delete("/channels/" + channelID + "/members/" + userID)
}

func (c *Client) GetUserChannels(teamID, userID string) ([]Channel, error) {
	var channels []Channel
	return channels, c.get("/users/"+userID+"/teams/"+teamID+"/channels", &channels)
}

// ─── Posts ───────────────────────────────────────────────────────────────────

func (c *Client) CreatePost(channelID, message string, fileIDs []string) (*Post, error) {
	body := map[string]interface{}{
		"channel_id": channelID,
		"message":    message,
	}
	if len(fileIDs) > 0 {
		body["file_ids"] = fileIDs
	}
	var p Post
	return &p, c.post("/posts", body, &p)
}

func (c *Client) GetChannelPosts(channelID string, since int64, perPage int) (*PostList, error) {
	path := fmt.Sprintf("/channels/%s/posts?per_page=%d", channelID, perPage)
	if since > 0 {
		path += fmt.Sprintf("&since=%d", since)
	}
	var pl PostList
	return &pl, c.get(path, &pl)
}

func (c *Client) GetChannelPostsPage(channelID string, page, perPage int) (*PostList, error) {
	path := fmt.Sprintf("/channels/%s/posts?page=%d&per_page=%d", channelID, page, perPage)
	var pl PostList
	return &pl, c.get(path, &pl)
}

// ─── Files ───────────────────────────────────────────────────────────────────

func (c *Client) UploadFile(channelID string, data []byte, filename, contentType string) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	_ = mw.WriteField("channel_id", channelID)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files"; filename="%s"`, filename))
	h.Set("Content-Type", contentType)
	fw, err := mw.CreatePart(h)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(data); err != nil {
		return "", err
	}
	mw.Close()

	req, err := http.NewRequest("POST", c.BaseURL+"/api/v4/files", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var mmErr mmError
		_ = json.NewDecoder(resp.Body).Decode(&mmErr)
		return "", fmt.Errorf("upload failed: %s (HTTP %d)", mmErr.Message, resp.StatusCode)
	}

	var result struct {
		FileInfos []FileInfo `json:"file_infos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.FileInfos) == 0 {
		return "", fmt.Errorf("no file info returned")
	}
	return result.FileInfos[0].ID, nil
}

// ─── System ──────────────────────────────────────────────────────────────────

func (c *Client) GetSystemStats() (map[string]interface{}, error) {
	var stats map[string]interface{}
	return stats, c.get("/system/analytics/old", &stats)
}

func (c *Client) ServerVersion() (string, error) {
	resp, err := c.do("GET", "/system/ping", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	ver := resp.Header.Get("X-Version-Id")
	if ver == "" {
		ver = "unknown"
	}
	return ver, nil
}
