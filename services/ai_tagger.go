package services

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/go-resty/resty/v2"
	youlog2 "github.com/project-xpolaris/youplustoolkit/youlog"
	"github.com/projectxpolaris/youcomic/config"
	"github.com/projectxpolaris/youcomic/plugin"
	"time"
)

var Logger *youlog2.Scope

var AiTaggerInstance *AiTaggerClient

func InitAiTaggerService() error {
	Logger = plugin.DefaultYouLogPlugin.Logger.NewScope("AiTagger")
	Logger.Info("initial ai tagger service...")
	AiTaggerInstance = NewAiTaggerClient()
	// check is accessable
	if AiTaggerInstance == nil {
		return nil
	}
	result, err := AiTaggerInstance.info()
	if err != nil {
		return err
	}
	if !result.Success {
		return errors.New("ai tagger service is not accessable")
	}
	return nil
}

type AiTaggerClient struct {
	Client *resty.Client
}

func NewAiTaggerClient() *AiTaggerClient {
	configManager := config.DefaultConfigProvider.Manager
	enable := configManager.GetBool("ai_tagger.enable")
	if !enable {
		Logger.Info("ai tagger service is disabled")
		return nil
	}
	Logger.Info("ai tagger service is enabled")
	url := configManager.GetString("ai_tagger.url")
	client := resty.New()
	client.SetBaseURL(url)
	client.SetTimeout(5 * time.Second)
	return &AiTaggerClient{
		Client: client,
	}
}

type TaggerItem struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

func (c *AiTaggerClient) predict(text string) ([]TaggerItem, error) {
	var responseBody []TaggerItem
	resp, err := c.Client.R().SetBody(haruka.JSON{
		"text": text,
	}).SetResult(&responseBody).Post("/predict")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, err
	}
	return responseBody, nil
}

// request for batch-predict

type MatchItem struct {
	Text   string       `json:"text"`
	Result []TaggerItem `json:"result"`
}

func (c *AiTaggerClient) batchPredict(texts []string) ([]MatchItem, error) {
	var responseBody []MatchItem
	resp, err := c.Client.R().SetBody(haruka.JSON{
		"texts": texts,
	}).SetResult(&responseBody).Post("/batch-predict")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, err
	}
	return responseBody, nil
}

type InfoBody struct {
	Success bool   `json:"success"`
	Input   string `json:"input"`
}

// request for /info

func (c *AiTaggerClient) info() (InfoBody, error) {
	var responseBody InfoBody
	resp, err := c.Client.R().SetResult(&responseBody).Get("/info")
	if err != nil {
		return InfoBody{}, err
	}
	if resp.IsError() {
		return InfoBody{}, err
	}
	return responseBody, nil
}
