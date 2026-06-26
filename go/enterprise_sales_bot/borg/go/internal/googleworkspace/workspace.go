package googleworkspace

import "fmt"

type Service struct {
	clientID     string
	clientSecret string
}

func NewService(clientID, clientSecret string) *Service {
	return &Service{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (s *Service) CreateDoc(title string) (string, error) {
	return fmt.Sprintf("Document '%s' created (placeholder)", title), nil
}

func (s *Service) AppendToDoc(docID, content string) error {
	return nil
}

func (s *Service) ReadDoc(docID string) (string, error) {
	return "[placeholder document content]", nil
}

func (s *Service) ListDocs() ([]string, error) {
	return []string{"doc-1", "doc-2"}, nil
}
