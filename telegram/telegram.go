package telegram

import (
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	Token  string
	ChatID string
	ApiURL string
}

func NewClient(token, chatID string) *Client {
	return &Client{
		Token:  token,
		ChatID: chatID,
		ApiURL: "https://api.telegram.org",
	}
}

func (c *Client) Send(message string) error {
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", c.ApiURL, c.Token)
	
	resp, err := http.PostForm(endpoint, url.Values{
		"chat_id":    {c.ChatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		return fmt.Errorf("post telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
