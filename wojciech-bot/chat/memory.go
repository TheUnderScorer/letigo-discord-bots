package chat

type ForgetButtonPayload struct {
	VectorStoreID string `json:"vector_store_id"`
	VectorFileID  string `json:"vector_file_id"`
	Content       string `json:"content"`
}

const ForgetButtonID = "forget"
