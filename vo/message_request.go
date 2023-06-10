/* This part of the code is reserved for functionality */
package vo

type ListMessages struct {
	PageNum  uint `json:"page_num"`
	PageSize uint `json:"page_size"`
	Type     uint `json:"type"`
}

type DeleteMessage struct {
	Ids []uint `json:"messageIds"`
}

type HaveReadMessage struct {
	Ids      []uint `json:"messageIds"`
	HaveRead bool   `json:"have_read"`
}
