package gpt

import "strings"

func GetLlmClientGptModels() map[string]*GPTModel {
	Models := make(map[string]*GPTModel)
	Models["gpt3Turbo"] = ModelGPT3Turbo
	Models["gpt3TurboBig"] = ModelGPT3TurboBig
	Models["gpt4"] = ModelGPT4
	Models["gpt4Big"] = ModelGPT4Big
	Models["gpt4Turbo"] = ModelGPT4Turbo
	Models["gpt4Vision"] = ModelGPT4Vision 
	return Models
}

func GetListOfModels() []string {
	modelsMap := GetLlmClientGptModels()
	keys := make([]string, 0, len(modelsMap))
	for key := range modelsMap {
		keys = append(keys, key)
	}
	return keys
}

func IsModelGPTValid(modelName string) bool {
	if notInList(modelName, GetListOfModels()) {
		return false
	}
	if !isModelGpt(modelName) {
		return false
	}
	return true
}

func notInList(target string, list []string) bool {
	for _, item := range list {
		if item == target {
			return false
		}
	}
	return true
}

func isModelGpt(modelName string) bool {
	contains := strings.Contains(strings.ToLower(modelName), "gpt")
	return contains
}
