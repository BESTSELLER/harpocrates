package lsp

type DocumentStore struct {
	documents map[string]string
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: make(map[string]string),
	}
}

func (s *DocumentStore) Open(uri string, text string) {
	s.documents[uri] = text
}

func (s *DocumentStore) Change(uri string, text string) {
	s.documents[uri] = text
}

func (s *DocumentStore) Get(uri string) (string, bool) {
	text, ok := s.documents[uri]
	return text, ok
}
