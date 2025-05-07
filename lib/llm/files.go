package llm

type File struct {
	Data        []byte
	Name        string
	ContentType string
}

func NewFile(data []byte, name string, contentType string) *File {
	return &File{
		Data:        data,
		Name:        name,
		ContentType: contentType,
	}
}
