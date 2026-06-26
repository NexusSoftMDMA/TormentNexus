package citation

import "fmt"

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) GenerateCitation(url, title string) string {
	return fmt.Sprintf("[%s](%s)", title, url)
}

func (s *Service) FormatCitations(citations []string) string {
	result := "\n### References\n"
	for i, c := range citations {
		result += fmt.Sprintf("%d. %s\n", i+1, c)
	}
	return result
}
