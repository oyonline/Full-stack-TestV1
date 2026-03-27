package dto

import (
	"encoding/json"
	"log"
	"net/http"
)

type FeishuResponseData struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data ResponseResult `json:"data"`
}

type Option struct {
	Id        string `json:"id"`
	Value     string `json:"value"`
	IsDefault bool   `json:"isDefault"`
}

type I18nResource struct {
	Location  string            `json:"location"`
	IsDefault bool              `json:"isDefault"`
	Texts     map[string]string `json:"texts"`
}

type ResponseData struct {
	Options       []Option       `json:"options"`
	I18nResources []I18nResource `json:"i18nResources"`
}

type ResponseResult struct {
	Result ResponseData `json:"result"`
}

func (e *FeishuResponseData) SetData(i18ns []I18nResource, options []Option) {
	response := ResponseResult{
		Result: ResponseData{
			Options:       options,
			I18nResources: i18ns,
		},
	}
	e.Msg = "success!"
	e.Data = response
}

func (e *FeishuResponseData) OK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	e.Code = 0
	e.Msg = "success!"
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(*e); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
