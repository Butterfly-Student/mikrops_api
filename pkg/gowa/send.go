package gowa

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// SendMessage sends a text message to a phone number or group.
// phone should be in format: 6289685028129@s.whatsapp.net (individual)
// or 120363347168689807@g.us (group)
func (c *Client) SendMessage(ctx context.Context, req SendMessageRequest) (*SendResponse, error) {
	return c.SendMessageWithDevice(ctx, req, "")
}

// SendMessageWithDevice sends a text message using the specified device.
// If deviceID is empty, the default device ID from config is used.
func (c *Client) SendMessageWithDevice(ctx context.Context, req SendMessageRequest, deviceID string) (*SendResponse, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPost, "/send/message", req, deviceID)
	if err != nil {
		return nil, fmt.Errorf("send message request failed: %w", err)
	}

	var result SendResponse
	if err := parseResponse(body, statusCode, &result); err != nil {
		return nil, fmt.Errorf("send message failed: %w", err)
	}

	return &result, nil
}

// SendTextMessage is a convenience method for sending a simple text message.
// phone should be in format: 6289685028129@s.whatsapp.net
func (c *Client) SendTextMessage(ctx context.Context, phone, message string) (*SendResponse, error) {
	return c.SendMessage(ctx, SendMessageRequest{
		Phone:   phone,
		Message: message,
	})
}

// SendGroupMessage is a convenience method for sending a message to a group.
// groupJID should be in format: 120363347168689807@g.us
func (c *Client) SendGroupMessage(ctx context.Context, groupJID, message string) (*SendResponse, error) {
	return c.SendMessage(ctx, SendMessageRequest{
		Phone:   groupJID,
		Message: message,
	})
}

// SendImageFromURL sends an image from a URL to a phone number or group.
func (c *Client) SendImageFromURL(ctx context.Context, phone, imageURL, caption string, deviceID string) (*SendResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("phone", phone)
	_ = writer.WriteField("image_url", imageURL)
	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}
	writer.Close()

	return c.sendMultipart(ctx, "/send/image", &buf, writer.FormDataContentType(), deviceID)
}

// SendImageFromFile sends an image from a local file to a phone number or group.
func (c *Client) SendImageFromFile(ctx context.Context, phone, filePath, caption string, deviceID string) (*SendResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("phone", phone)
	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	part, err := writer.CreateFormFile("image", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	writer.Close()

	return c.sendMultipart(ctx, "/send/image", &buf, writer.FormDataContentType(), deviceID)
}

// SendFileFromURL sends a file from a URL to a phone number or group.
func (c *Client) SendFileFromURL(ctx context.Context, phone, fileURL, caption string, deviceID string) (*SendResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("phone", phone)
	_ = writer.WriteField("file_url", fileURL)
	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}
	writer.Close()

	return c.sendMultipart(ctx, "/send/file", &buf, writer.FormDataContentType(), deviceID)
}

// SendVideoFromURL sends a video from a URL to a phone number or group.
func (c *Client) SendVideoFromURL(ctx context.Context, phone, videoURL, caption string, deviceID string) (*SendResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("phone", phone)
	_ = writer.WriteField("video_url", videoURL)
	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}
	writer.Close()

	return c.sendMultipart(ctx, "/send/video", &buf, writer.FormDataContentType(), deviceID)
}

// sendMultipart sends a multipart form request to the Gowa API.
func (c *Client) sendMultipart(ctx context.Context, path string, body *bytes.Buffer, contentType string, deviceID string) (*SendResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Password)
	req.Header.Set("Content-Type", contentType)

	if deviceID != "" {
		req.Header.Set("X-Device-Id", deviceID)
	} else if c.config.DeviceID != "" {
		req.Header.Set("X-Device-Id", c.config.DeviceID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result SendResponse
	if err := parseResponse(respBody, resp.StatusCode, &result); err != nil {
		return nil, fmt.Errorf("send multipart failed: %w", err)
	}

	return &result, nil
}
