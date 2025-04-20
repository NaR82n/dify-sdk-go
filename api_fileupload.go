package dify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

type FileUploadRequest struct {
	FilePath string `json:"file"`
	User     string `json:"user"`
}

type FileUploadResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`
	CreatedBy string `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
}

func (api *API) UploadFile(ctx context.Context, request FileUploadRequest) (*FileUploadResponse, error) {

	file, err := os.Open(request.FilePath)
	if err != nil {
		log.Fatalf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 创建一个新的 multipart 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", request.FilePath)
	if err != nil {
		log.Fatalf("创建文件字段失败: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatalf("复制文件内容失败: %v", err)
	}

	// 添加用户字段
	err = writer.WriteField("user", request.User)
	if err != nil {
		log.Fatalf("添加用户字段失败: %v", err)
	}

	// 关闭 writer
	err = writer.Close()
	if err != nil {
		log.Fatalf("关闭 writer 失败: %v", err)
	}

	req, err := api.createBaseRequestRaw(ctx, http.MethodPost, "/v1/files/upload",
		writer.FormDataContentType(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create base request: %w", err)
	}

	resp, err := api.c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %s: %s", resp.Status, readResponseBody(resp.Body))
	}

	var fileuploadResp FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileuploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &fileuploadResp, nil
}
