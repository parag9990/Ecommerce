package dto

import "product-service/internal/usecase"

type CollectionSetupResponseDTO struct {
	Database    string                   `json:"database"`
	Collections []CollectionSetupItemDTO `json:"collections"`
}

type CollectionSetupItemDTO struct {
	Name             string   `json:"name"`
	Created          bool     `json:"created"`
	ValidatorApplied bool     `json:"validator_applied"`
	IndexNames       []string `json:"index_names"`
}

type CollectionDescriptionResponseDTO struct {
	Collections []CollectionDescriptionDTO `json:"collections"`
}

type CollectionDescriptionDTO struct {
	Name         string   `json:"name"`
	IndexNames   []string `json:"index_names"`
	HasValidator bool     `json:"has_validator"`
}

func CollectionSetupResponseFromUseCase(result usecase.CollectionSetupResult) CollectionSetupResponseDTO {
	response := CollectionSetupResponseDTO{
		Database:    result.Database,
		Collections: make([]CollectionSetupItemDTO, 0, len(result.Collections)),
	}
	for _, item := range result.Collections {
		response.Collections = append(response.Collections, CollectionSetupItemDTO{
			Name:             item.Name,
			Created:          item.Created,
			ValidatorApplied: item.ValidatorApplied,
			IndexNames:       append([]string(nil), item.IndexNames...),
		})
	}
	return response
}

func CollectionDescriptionResponseFromUseCase(descriptions []usecase.CollectionDescription) CollectionDescriptionResponseDTO {
	response := CollectionDescriptionResponseDTO{
		Collections: make([]CollectionDescriptionDTO, 0, len(descriptions)),
	}
	for _, description := range descriptions {
		response.Collections = append(response.Collections, CollectionDescriptionDTO{
			Name:         description.Name,
			IndexNames:   append([]string(nil), description.IndexNames...),
			HasValidator: description.HasValidator,
		})
	}
	return response
}
