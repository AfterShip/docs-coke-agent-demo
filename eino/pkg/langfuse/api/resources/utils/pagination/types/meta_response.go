package types

// MetaResponse represents pagination metadata in API responses
type MetaResponse struct {
	Page         int  `json:"page"`
	Limit        int  `json:"limit"`
	TotalItems   int  `json:"totalItems"`
	TotalPages   int  `json:"totalPages"`
	HasNextPage  bool `json:"hasNextPage"`
	HasPrevPage  bool `json:"hasPrevPage"`
}

// NewMetaResponse creates a new meta response for pagination
func NewMetaResponse(page, limit, totalItems int) *MetaResponse {
	totalPages := (totalItems + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	
	return &MetaResponse{
		Page:         page,
		Limit:        limit,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
	}
}

// GetNextPage returns the next page number if it exists
func (m *MetaResponse) GetNextPage() *int {
	if !m.HasNextPage {
		return nil
	}
	nextPage := m.Page + 1
	return &nextPage
}

// GetPrevPage returns the previous page number if it exists
func (m *MetaResponse) GetPrevPage() *int {
	if !m.HasPrevPage {
		return nil
	}
	prevPage := m.Page - 1
	return &prevPage
}

// GetFirstPage returns the first page number
func (m *MetaResponse) GetFirstPage() int {
	return 1
}

// GetLastPage returns the last page number
func (m *MetaResponse) GetLastPage() int {
	return m.TotalPages
}

// IsFirstPage returns true if this is the first page
func (m *MetaResponse) IsFirstPage() bool {
	return m.Page == 1
}

// IsLastPage returns true if this is the last page
func (m *MetaResponse) IsLastPage() bool {
	return m.Page == m.TotalPages
}

// GetOffset returns the offset for the current page
func (m *MetaResponse) GetOffset() int {
	return (m.Page - 1) * m.Limit
}

// GetItemRange returns the range of items on the current page
func (m *MetaResponse) GetItemRange() (int, int) {
	start := m.GetOffset() + 1
	end := start + m.Limit - 1
	
	if end > m.TotalItems {
		end = m.TotalItems
	}
	
	if start > m.TotalItems {
		start = m.TotalItems
	}
	
	return start, end
}