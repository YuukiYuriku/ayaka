package pagination

//parameter for pagination
type PaginationParam struct {
    Page     int `json:"page"`
    PageSize int `json:"page_size"`
}

//response pagination
type PaginationResponse struct {
    Data         interface{}        `json:"results"`
    TotalRecords int                `json:"total_records"`
    TotalPages   int                `json:"total_pages"`
    CurrentPage  int                `json:"current_page"`
    PageSize     int                `json:"page_size"`
    HasNext      bool               `json:"has_next"`
    HasPrevious  bool               `json:"has_previous"`
}

func CountPagination(param *PaginationParam, totalRecords int) (int, int) {
	// count total pages
    totalPages := (totalRecords + param.PageSize - 1) / param.PageSize

    // count offset
    offset := (param.Page - 1) * param.PageSize

	return totalPages, offset
}