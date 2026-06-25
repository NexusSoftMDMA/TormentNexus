package tools

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ArXiv API response structures
type ArXivFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []ArXivEntry `xml:"entry"`
}

type ArXivEntry struct {
	XMLName     xml.Name   `xml:"entry"`
	Title       string     `xml:"title"`
	Summary     string     `xml:"summary"`
	Author      []Author   `xml:"author"`
	Published   string     `xml:"published"`
	Updated     string     `xml:"updated"`
	Link        []Link     `xml:"link"`
	Category    []Category `xml:"category"`
	ID          string     `xml:"id"`
	Doi         string     `xml:"arxiv:doi,omitempty"`
	Comment     string     `xml:"arxiv:comment,omitempty"`
	JournalRef  string     `xml:"arxiv:journal_ref,omitempty"`
}

type Author struct {
	Name string `xml:"name"`
}

type Link struct {
	Href  string `xml:"href,attr"`
	Title string `xml:"title,attr"`
	Type  string `xml:"type,attr"`
	Rel   string `xml:"rel,attr"`
}

type Category struct {
	Term string `xml:"term,attr"`
}

// cleanText removes extra whitespace from text
func cleanText(s string) string {
	// Replace multiple whitespace characters with a single space
	var b strings.Builder
	inSpace := false
	for _, r := range s {
		if strings.IsSpace(r) {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}

// HandleSearchPapers searches for academic papers on arXiv
func HandleSearchPapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	maxResults, _ :=getInt(args, "max_results")
	if maxResults == 0 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50
	}

	start, _ :=getInt(args, "start")
	if start < 0 {
		start = 0
	}

	searchQuery := url.Values{}
	searchQuery.Set("search_query", "all:"+query)
	searchQuery.Set("start", strconv.Itoa(start))
	searchQuery.Set("max_results", strconv.Itoa(maxResults))
	searchQuery.Set("sortBy", "relevance")

	apiURL := "http://export.arxiv.org/api/query?" + searchQuery.Encode()

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("arXiv API error: %d - %s", resp.StatusCode, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var feed ArXivFeed
	parseErr := xml.Unmarshal(body, &feed)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(feed.Entries) == 0 {
		return ok("No papers found for query: " + query)
}

	var results []string
	results = append(results, fmt.Sprintf("Found %d papers for query: %s\n", len(feed.Entries), query))
	results = append(results, "="+strings.Repeat("=", 60)+"\n\n")

	for i, entry := range feed.Entries {
		results = append(results, fmt.Sprintf("[%d] %s", start+i+1, cleanText(entry.Title)))
		results = append(results, "")

		authors := make([]string, len(entry.Author))
		for j, a := range entry.Author {
			authors[j] = a.Name
		}
		results = append(results, fmt.Sprintf("Authors: %s", strings.Join(authors, ", ")))
		
		// Fix potential panic by checking length
		publishedDate := entry.Published
		if len(publishedDate) >= 10 {
			publishedDate = publishedDate[:10]
		}
		results = append(results, fmt.Sprintf("Published: %s", publishedDate))
		results = append(results, fmt.Sprintf("ID: %s", extractArxivID(entry.ID)))
		results = append(results, "")

		summary := cleanText(entry.Summary)
		if len(summary) > 300 {
			summary = summary[:300] + "..."
		}
		results = append(results, "Summary: "+summary)
		results = append(results, "")

		if len(entry.Category) > 0 {
			cats := make([]string, len(entry.Category))
			for j, c := range entry.Category {
				cats[j] = c.Term
			}
			results = append(results, fmt.Sprintf("Categories: %s", strings.Join(cats, ", ")))

		results = append(results, fmt.Sprintf("URL: %s", entry.ID))
		results = append(results, "")
		results = append(results, "-"+strings.Repeat("-", 60))
		results = append(results, "")

	return ok(strings.Join(results, "\n"))
}

}
}

// HandleGetPaper retrieves detailed information about a specific paper
func HandleGetPaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ :=getString(args, "paper_id")
	if paperID == "" {
		return err("paper_id parameter is required")
}

	arxivID := extractArxivID(paperID)
	if arxivID == "" {
		return err("invalid paper ID format")
}

	apiURL := "http://export.arxiv.org/api/query?id_list=" + arxivID

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("arXiv API error: %d - %s", resp.StatusCode, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var feed ArXivFeed
	parseErr := xml.Unmarshal(body, &feed)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(feed.Entries) == 0 {
		return err("Paper not found: " + arxivID)
}

	entry := feed.Entries[0]
	var details []string

	details = append(details, "="+strings.Repeat("=", 60))
	details = append(details, "PAPER DETAILS")
	details = append(details, "="+strings.Repeat("=", 60))
	details = append(details, "")
	details = append(details, "Title: "+cleanText(entry.Title))
	details = append(details, "")

	authors := make([]string, len(entry.Author))
	for i, a := range entry.Author {
		authors[i] = a.Name
	}
	details = append(details, "Authors: "+strings.Join(authors, ", "))
	details = append(details, "")

	details = append(details, "Published: "+entry.Published)
	details = append(details, "Updated: "+entry.Updated)
	details = append(details, "arXiv ID: "+extractArxivID(entry.ID))
	details = append(details, "")

	details = append(details, "Abstract:")
	details = append(details, cleanText(entry.Summary))
	details = append(details, "")

	if len(entry.Category) > 0 {
		cats := make([]string, len(entry.Category))
		for i, c := range entry.Category {
			cats[i] = c.Term
		}
		details = append(details, "Categories: "+strings.Join(cats, ", "))

	if entry.Doi != "" {
		details = append(details, "DOI: "+entry.Doi)

	if entry.Comment != "" {
		details = append(details, "")
		details = append(details, "Comments: "+cleanText(entry.Comment))

	if entry.JournalRef != "" {
		details = append(details, "Journal Ref: "+entry.JournalRef)

	details = append(details, "")
	details = append(details, "PDF URL: https://arxiv.org/pdf/"+extractArxivID(entry.ID)+".pdf")
	details = append(details, "arXiv URL: "+entry.ID)

	return ok(strings.Join(details, "\n"))
}

}
}
}
}

// HandleGetRecentPapers retrieves recent papers by category
func HandleGetRecentPapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	maxResults, _ :=getInt(args, "max_results")
	if maxResults == 0 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50
	}

	searchQuery := url.Values{}
	if category != "" {
		searchQuery.Set("search_query", "cat:"+category)
	} else {
		searchQuery.Set("search_query", "all:*")

	searchQuery.Set("start", "0")
	searchQuery.Set("max_results", strconv.Itoa(maxResults))
	searchQuery.Set("sortBy", "submittedDate")
	searchQuery.Set("sortOrder", "descending")

	apiURL := "http://export.arxiv.org/api/query?" + searchQuery.Encode()

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("arXiv API error: %d - %s", resp.StatusCode, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var feed ArXivFeed
	parseErr := xml.Unmarshal(body, &feed)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(feed.Entries) == 0 {
		catInfo := "all categories"
		if category != "" {
			catInfo = "category: " + category
		}
		return ok("No recent papers found for " + catInfo)
}

	var results []string
	catInfo := "all categories"
	if category != "" {
		catInfo = "category: " + category
	}
	results = append(results, fmt.Sprintf("Recent papers for %s\n", catInfo))
	results = append(results, "="+strings.Repeat("=", 60)+"\n\n")

	for i, entry := range feed.Entries {
		results = append(results, fmt.Sprintf("[%d] %s", i+1, cleanText(entry.Title)))
		results = append(results, "")

		authors := make([]string, len(entry.Author))
		for j, a := range entry.Author {
			authors[j] = a.Name
		}
		results = append(results, fmt.Sprintf("Authors: %s", strings.Join(authors, ", ")))
		
		// Fix potential panic by checking length
		publishedDate := entry.Published
		if len(publishedDate) >= 10 {
			publishedDate = publishedDate[:10]
		}
		results = append(results, fmt.Sprintf("Date: %s", publishedDate))
		results = append(results, fmt.Sprintf("arXiv: %s", extractArxivID(entry.ID)))

		if len(entry.Category) > 0 {
			results = append(results, fmt.Sprintf("Category: %s", entry.Category[0].Term))

		results = append(results, "")

	return ok(strings.Join(results, "\n"))
}

}
}
}

// HandleSearchByAuthor searches for papers by author name
func HandleSearchByAuthor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	author, _ :=getString(args, "author")
	if author == "" {
		return err("author parameter is required")
}

	maxResults, _ :=getInt(args, "max_results")
	if maxResults == 0 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50
	}

	searchQuery := url.Values{}
	searchQuery.Set("search_query", "au:"+author)
	searchQuery.Set("start", "0")
	searchQuery.Set("max_results", strconv.Itoa(maxResults))
	searchQuery.Set("sortBy", "relevance")

	apiURL := "http://export.arxiv.org/api/query?" + searchQuery.Encode()

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("arXiv API error: %d - %s", resp.StatusCode, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var feed ArXivFeed
	parseErr := xml.Unmarshal(body, &feed)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if len(feed.Entries) == 0 {
		return ok("No papers found for author: " + author)
}

	var results []string
	results = append(results, fmt.Sprintf("Papers by author: %s\n", author))
	results = append(results, "="+strings.Repeat("=", 60)+"\n\n")

	for i, entry := range feed.Entries {
		results = append(results, fmt.Sprintf("[%d] %s", i+1, cleanText(entry.Title)))
		results = append(results, "")

		authors := make([]string, len(entry.Author))
		for j, a := range entry.Author {
			authors[j] = a.Name
		}
		results = append(results, fmt.Sprintf("All Authors: %s", strings.Join(authors, ", ")))
		
		// Fix potential panic by checking length
		publishedDate := entry.Published
		if len(publishedDate) >= 10 {
			publishedDate = publishedDate[:10]
		}
		results = append(results, fmt.Sprintf("Published: %s", publishedDate))
		results = append(results, fmt.Sprintf("arXiv ID: %s", extractArxivID(entry.ID)))
		results = append(results, "")

	return ok(strings.Join(results, "\n"))
}

}

// extractArxivID extracts the arXiv ID from a full URL
func extractArxivID(idURL string) string {
	parts := strings.Split(idURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return idURL
}

// HandleListCategories lists available arXiv categories
func HandleListCategories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Common arXiv categories organized by field
	categories := map[string][]string{
		"Computer Science": {
			"cs.AI - Artificial Intelligence",
			"cs.CL - Computation and Language",
			"cs.CV - Computer Vision and Pattern Recognition",
			"cs.LG - Machine Learning",
			"cs.NE - Neural and Evolutionary Computing",
			"cs.RO - Robotics",
			"cs.SE - Software Engineering",
			"cs.IR - Information Retrieval",
			"cs.HC - Human-Computer Interaction",
			"cs.CR - Cryptography and Security",
		},
		"Physics": {
			"astro-ph - Astrophysics",
			"cond-mat - Condensed Matter",
			"gr-qc - General Relativity and Quantum Cosmology",
			"hep-ex - High Energy Physics - Experiment",
			"hep-lat - High Energy Physics - Lattice",
			"hep-ph - High Energy Physics - Phenomenology",
			"hep-th - High Energy Physics - Theory",
			"math-ph - Mathematical Physics",
			"nlin - Nonlinear Sciences",
			"nucl-ex - Nuclear Experiment",
			"nucl-th - Nuclear Theory",
			"physics.acc-ph - Accelerator Physics",
			"physics.app-ph - Applied Physics",
			"physics.atom-ph - Atomic, Molecular and Optical Physics",
			"physics.bio-ph - Biological Physics",
			"physics.chem-ph - Chemical Physics",
			"physics.class-ph - Classical Physics",
			"physics.comp-ph - Computational Physics",
			"physics.data-an - Data Analysis",
			"physics.ed-ph - Physics Education",
			"physics.flu-dyn - Fluid Dynamics",
			"physics.gen-ph - General Physics",
			"physics.geo-ph - Geophysics",
			"physics.hist-ph - History of Physics",
			"physics.ins-det - Instrumentation",
			"physics.med-ph - Medical Physics",
			"physics.optics - Optics",
			"physics.plasm-ph - Plasma Physics",
		}, // Added missing comma here
	}

	return ok("not yet implemented")
}