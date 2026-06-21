package letterboxd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Todo add Mock Roundtripper to test Movies() iterator

type diaryTestCase struct {
	name     string
	filename string
	validate func(*testing.T, []*DiaryEntry, string)
}

func TestParseDiary(t *testing.T) {
	tests := []diaryTestCase{
		{
			name:     "AbbyNormalz Diary",
			filename: "abbynormalz.html",
			validate: validateDiaryAbbyNormalz,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "diaries", tt.filename)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to read test data %s: %v", tt.filename, err)
			}
			doc, err := getDocument(f)
			if err != nil {
				t.Fatalf("failed to parse test data doc: %v", err)
			}
			entries, nextLink := parseDiaryEntries(doc)
			tt.validate(t, entries, nextLink)

		})
	}

}

func validateDiaryAbbyNormalz(t *testing.T, actual []*DiaryEntry, nextLink string) {
	if len(actual) != 50 {
		t.Errorf("parseDiaryEntries() length expected '20', got %d", len(actual))
	}
	if len(actual) == 50 {
		entry := actual[0]
		expectedTimeWatched := time.Date(2026, time.June, 20, 0, 0, 0, 0, time.UTC)
		if entry.DateWatched.Format("2006-01-02") != expectedTimeWatched.Format("2006-01-02") {
			t.Errorf("parseDiaryEntries()[0] DateWatched expected `2026-06-21'`, received '%s", entry.DateWatched.Format("2006-01-02"))
		}

		if entry.Rating != 7 {
			t.Errorf("parseDiaryEntries()[0] Rating expected '7', got '%d'", entry.Rating)
		}

		if entry.ReleaseYear != 2026 {
			t.Errorf("parseDiaryEntries()[0] ReleaseYear expected '2026', got '%d'", entry.ReleaseYear)
		}

		if entry.Slug != "toy-story-5" {
			t.Errorf("parseDiaryEntries()[0] Slug expected 'toy-story-5', got '%s'", entry.Slug)
		}

		if entry.Title != "Toy Story 5" {
			t.Errorf("parseDiaryEntries()[0] Title expected 'Toy Story 5', got '%s'", entry.Title)
		}
	}

}
