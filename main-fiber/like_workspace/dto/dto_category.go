package dto

// เพื่อใช้กับ "รายการโพสต์" (generic รองรับชนิดใดก็ได้)
type ListByCategoryResp[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
	TotalCount int64   `json:"total_count"`
}
