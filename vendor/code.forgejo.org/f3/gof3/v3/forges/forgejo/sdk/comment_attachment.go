// Copyright F3 Authors
// SPDX-License-Identifier: MIT

package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func (c *Client) ListCommentAttachments(user, repo string, issue int64) ([]*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	attachments := make([]*Attachment, 0)
	resp, err := c.getParsedResponse("GET",
		fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets", user, repo, issue),
		nil, nil, &attachments)
	return attachments, resp, err
}

func (c *Client) GetCommentAttachment(user, repo string, issue, id int64) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	a := new(Attachment)
	resp, err := c.getParsedResponse("GET",
		fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets/%d", user, repo, issue, id),
		nil, nil, &a)
	return a, resp, err
}

func (c *Client) CreateCommentAttachment(user, repo string, issue int64, file io.Reader, filename string) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	// Write file to body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("attachment", filename)
	if err != nil {
		return nil, nil, err
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, nil, err
	}

	// Send request
	attachment := new(Attachment)
	resp, err := c.getParsedResponse("POST",
		fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets", user, repo, issue),
		http.Header{"Content-Type": {writer.FormDataContentType()}}, body, &attachment)
	return attachment, resp, err
}

func (c *Client) EditCommentAttachment(user, repo string, issue, attachment int64, form EditAttachmentOptions) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&form)
	if err != nil {
		return nil, nil, err
	}
	attach := new(Attachment)
	resp, err := c.getParsedResponse("PATCH", fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets/%d", user, repo, issue, attachment), jsonHeader, bytes.NewReader(body), attach)
	return attach, resp, err
}

func (c *Client) DeleteCommentAttachment(user, repo string, issue, id int64) (*Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets/%d", user, repo, issue, id), nil, nil)
	return resp, err
}
