package httptransport

import (
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/usecase"
)

type schemaContractResponse struct {
	Version             string          `json:"version"`
	Collection          collectionDTO   `json:"collection"`
	TypesenseDefinition map[string]any  `json:"typesense_definition"`
	SearchPolicy        searchPolicyDTO `json:"search_policy"`
	SynonymModel        synonymModelDTO `json:"synonym_model"`
}

type collectionDTO struct {
	Name                string     `json:"name"`
	Fields              []fieldDTO `json:"fields"`
	DefaultSortingField string     `json:"default_sorting_field"`
}

type fieldDTO struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Facet    bool   `json:"facet,omitempty"`
	Sort     bool   `json:"sort,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

type searchPolicyDTO struct {
	CollectionName   string               `json:"collection_name"`
	Searchable       []searchableFieldDTO `json:"searchable_fields"`
	FacetFields      []string             `json:"facet_fields"`
	SortOptions      []sortOptionDTO      `json:"sort_options"`
	DefaultSort      string               `json:"default_sort"`
	DefaultFacetBy   string               `json:"default_facet_by"`
	DefaultQueryBy   string               `json:"default_query_by"`
	QueryByWeights   string               `json:"query_by_weights"`
	NumTypos         string               `json:"num_typos"`
	Prefix           string               `json:"prefix"`
	DefaultPage      int                  `json:"default_page"`
	DefaultPageSize  int                  `json:"default_page_size"`
	AllowedPageSizes []int                `json:"allowed_page_sizes"`
	APIContract      apiContractDTO       `json:"api_contract"`
}

type searchableFieldDTO struct {
	Name     string `json:"name"`
	Weight   int    `json:"weight"`
	NumTypos int    `json:"num_typos"`
	Prefix   bool   `json:"prefix"`
}

type sortOptionDTO struct {
	Key         string `json:"key"`
	TypesenseBy string `json:"typesense_by"`
	Description string `json:"description"`
}

type apiContractDTO struct {
	RESTMethod     string `json:"rest_method"`
	RESTPath       string `json:"rest_path"`
	GRPCMethod     string `json:"grpc_method"`
	RequestSchema  string `json:"request_schema"`
	ResponseSchema string `json:"response_schema"`
}

type synonymModelDTO struct {
	Shape    synonymDTO   `json:"shape"`
	Examples []synonymDTO `json:"examples"`
}

type synonymDTO struct {
	Root     string   `json:"root"`
	Synonyms []string `json:"synonyms"`
}

type createSynonymRequest struct {
	Root     string   `json:"root"`
	Synonyms []string `json:"synonyms"`
}

type reindexRequest struct {
	Mode             string `json:"mode"`
	BatchSize        int    `json:"batch_size"`
	Reason           string `json:"reason"`
	DryRun           bool   `json:"dry_run"`
	TargetCollection string `json:"target_collection"`
}

type reindexAcceptedResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type searchSynonymResponse struct {
	SynonymID string   `json:"synonym_id"`
	Root      string   `json:"root"`
	Synonyms  []string `json:"synonyms"`
}

type searchSynonymListResponse struct {
	Synonyms []searchSynonymResponse `json:"synonyms"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type searchProductsResponse struct {
	Products []productDTO               `json:"products"`
	Facets   map[string][]facetValueDTO `json:"facets"`
	Total    int                        `json:"total"`
}

type autocompleteResponse struct {
	Suggestions []string `json:"suggestions"`
}

type productDTO struct {
	ProductID   string              `json:"product_id"`
	SellerID    string              `json:"seller_id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Brand       string              `json:"brand"`
	CategoryID  string              `json:"category_id"`
	Status      string              `json:"status"`
	Variants    []productVariantDTO `json:"variants"`
}

type productVariantDTO struct {
	SKU           string         `json:"sku"`
	Attributes    map[string]any `json:"attributes,omitempty"`
	Price         *moneyDTO      `json:"price,omitempty"`
	StockQuantity int            `json:"stock_quantity"`
}

type moneyDTO struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type facetValueDTO struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

func schemaContractDTO(contract usecase.ProductSchemaContract) schemaContractResponse {
	fields := make([]fieldDTO, 0, len(contract.Collection.Fields))
	for _, field := range contract.Collection.Fields {
		fields = append(fields, fieldDTO{
			Name:     field.Name,
			Type:     field.Type,
			Facet:    field.Facet,
			Sort:     field.Sort,
			Optional: field.Optional,
		})
	}

	searchable := make([]searchableFieldDTO, 0, len(contract.SearchPolicy.Searchable))
	for _, field := range contract.SearchPolicy.Searchable {
		searchable = append(searchable, searchableFieldDTO{
			Name:     field.Name,
			Weight:   field.Weight,
			NumTypos: field.NumTypos,
			Prefix:   field.Prefix,
		})
	}

	sortOptions := make([]sortOptionDTO, 0, len(contract.SearchPolicy.SortOptions))
	for _, option := range contract.SearchPolicy.SortOptions {
		sortOptions = append(sortOptions, sortOptionDTO{
			Key:         option.Key,
			TypesenseBy: option.TypesenseBy,
			Description: option.Description,
		})
	}

	examples := make([]synonymDTO, 0, len(contract.SynonymModel.Examples))
	for _, synonym := range contract.SynonymModel.Examples {
		examples = append(examples, synonymDTO{
			Root:     synonym.Root,
			Synonyms: synonym.Synonyms,
		})
	}

	return schemaContractResponse{
		Version: contract.Version,
		Collection: collectionDTO{
			Name:                contract.Collection.Name,
			Fields:              fields,
			DefaultSortingField: contract.Collection.DefaultSortingField,
		},
		TypesenseDefinition: contract.TypesenseDefinition,
		SearchPolicy: searchPolicyDTO{
			CollectionName:   contract.SearchPolicy.CollectionName,
			Searchable:       searchable,
			FacetFields:      contract.SearchPolicy.FacetFields,
			SortOptions:      sortOptions,
			DefaultSort:      contract.SearchPolicy.DefaultSort,
			DefaultFacetBy:   contract.SearchPolicy.DefaultFacetBy,
			DefaultQueryBy:   contract.SearchPolicy.DefaultQueryBy,
			QueryByWeights:   contract.SearchPolicy.QueryByWeights,
			NumTypos:         contract.SearchPolicy.NumTypos,
			Prefix:           contract.SearchPolicy.Prefix,
			DefaultPage:      contract.SearchPolicy.DefaultPage,
			DefaultPageSize:  contract.SearchPolicy.DefaultPageSize,
			AllowedPageSizes: contract.SearchPolicy.AllowedPageSizes,
			APIContract: apiContractDTO{
				RESTMethod:     contract.SearchPolicy.APIContract.RESTMethod,
				RESTPath:       contract.SearchPolicy.APIContract.RESTPath,
				GRPCMethod:     contract.SearchPolicy.APIContract.GRPCMethod,
				RequestSchema:  contract.SearchPolicy.APIContract.RequestSchema,
				ResponseSchema: contract.SearchPolicy.APIContract.ResponseSchema,
			},
		},
		SynonymModel: synonymModelDTO{
			Shape: synonymDTO{
				Root:     contract.SynonymModel.Shape.Root,
				Synonyms: contract.SynonymModel.Shape.Synonyms,
			},
			Examples: examples,
		},
	}
}

func searchResponseDTO(resp domain.SearchResponse) searchProductsResponse {
	products := make([]productDTO, 0, len(resp.Products))
	for _, product := range resp.Products {
		variants := make([]productVariantDTO, 0, len(product.Variants))
		for _, variant := range product.Variants {
			var price *moneyDTO
			if variant.Price != nil {
				price = &moneyDTO{
					Amount:   variant.Price.Amount,
					Currency: variant.Price.Currency,
				}
			}
			variants = append(variants, productVariantDTO{
				SKU:           variant.SKU,
				Attributes:    variant.Attributes,
				Price:         price,
				StockQuantity: variant.StockQuantity,
			})
		}
		products = append(products, productDTO{
			ProductID:   product.ProductID,
			SellerID:    product.SellerID,
			Title:       product.Title,
			Description: product.Description,
			Brand:       product.Brand,
			CategoryID:  product.CategoryID,
			Status:      product.Status,
			Variants:    variants,
		})
	}

	facets := make(map[string][]facetValueDTO, len(resp.Facets))
	for field, values := range resp.Facets {
		facets[field] = make([]facetValueDTO, 0, len(values))
		for _, value := range values {
			facets[field] = append(facets[field], facetValueDTO{
				Value: value.Value,
				Count: value.Count,
			})
		}
	}

	return searchProductsResponse{
		Products: products,
		Facets:   facets,
		Total:    resp.Total,
	}
}

func autocompleteResponseDTO(resp domain.AutocompleteResponse) autocompleteResponse {
	suggestions := append([]string(nil), resp.Suggestions...)
	if suggestions == nil {
		suggestions = []string{}
	}
	return autocompleteResponse{Suggestions: suggestions}
}

func searchSynonymResponseDTO(synonym domain.SearchSynonym) searchSynonymResponse {
	synonyms := append([]string(nil), synonym.Synonyms...)
	if synonyms == nil {
		synonyms = []string{}
	}
	return searchSynonymResponse{
		SynonymID: synonym.ID,
		Root:      synonym.Root,
		Synonyms:  synonyms,
	}
}

func reindexAcceptedResponseDTO(accepted domain.ReindexAccepted) reindexAcceptedResponse {
	return reindexAcceptedResponse{
		JobID:  accepted.JobID,
		Status: accepted.Status,
	}
}

func searchSynonymListResponseDTO(synonyms []domain.SearchSynonym) searchSynonymListResponse {
	items := make([]searchSynonymResponse, 0, len(synonyms))
	for _, synonym := range synonyms {
		items = append(items, searchSynonymResponseDTO(synonym))
	}
	return searchSynonymListResponse{Synonyms: items}
}
