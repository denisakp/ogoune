package dto

import "github.com/denisakp/ogoune/internal/domain"

type CreateTagDto struct {
	Name        string  `json:"name" binding:"required"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
}

type UpdateTagDto struct {
	Name        *string `json:"name,omitempty"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
}

type TagResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
}

func ToTagResponse(tag *domain.Tags) *TagResponse {
	return &TagResponse{
		ID:          tag.ID,
		Name:        tag.Name,
		Color:       tag.Color,
		Description: tag.Description,
	}
}
