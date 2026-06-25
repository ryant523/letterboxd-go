package letterboxd

import "strings"

// Rating are the letterboxd ratings that can be used to filter diary results
type Rating string

const (
	RatingHalf      Rating = "0.5"
	RatingOne       Rating = "1"
	RatingOneHalf   Rating = "1.5"
	RatingTwo       Rating = "2"
	RatingTwoHalf   Rating = "2.5"
	RatingThree     Rating = "3"
	RatingThreeHalf Rating = "3.5"
	RatingFour      Rating = "4"
	RatingFourHalf  Rating = "4.5"
	RatingFive      Rating = "5"
)

// String satisfies the fmt.Stringer interface
func (r Rating) String() string {
	return string(r)
}

// Sort Options used for lists and diaries
type SortBy string

const (
	SortAddedNewest     SortBy = "by/added"
	SortAddedEarliest   SortBy = "by/added-earliest"
	SortReleaseNewest   SortBy = "by/release"
	SortReleaseEarliest SortBy = "by/release-earliest"
	SortFilmName        SortBy = "by/name"
	SortPopularity      SortBy = "by/popular"
	SortRatingHighest   SortBy = "by/rating"
	SortRatingLowest    SortBy = "by/rating-lowest"
	SortShortestLength  SortBy = "by/shortest"
	SortLongestLength   SortBy = "by/longest"

	// Diary Specific
	SortEntryRatingHighest SortBy = "by/entry-rating"
	SortEntryRatingLowest  SortBy = "by/entry-rating-lowest"
	SortByActivity         SortBy = "by/activity"
	SortDiaryCount         SortBy = "by/diary-count"

	// List Specific
	SortListOwnerRatingHighest SortBy = "by/owner-rating"
	SortListOwnerRatingLowest  SortBy = "by/ower-rating-lowest"
	SortListShuffle            SortBy = "by/shuffle"
	SortListReverse            SortBy = "by/reverse"
	SortListOwnerDiaryNewest   SortBy = "by/owner-diary"
	SortListOwnerDiaryEarliest SortBy = "by/owner-diry-earliest"
)

func (s SortBy) String() string {
	return string(s)
}

// Options are build the filter/sort options in the url
type Options struct {
	Sort        SortBy
	WatchedYear string
	Genre       string
	Decade      string
	Year        string
	Director    string
	Actor       string
	Rating      Rating
}

// Option is a functional option used to add the filter/sort options in the url
type Option func(*Options)

// WithSortBy will sort the diary and list by the specified SortBy
func WithSortBy(sort SortBy) Option {
	return func(c *Options) { c.Sort = sort }
}

// WithWatchedYear will filter the diary by the year watched
func WithWatchedYear(year string) Option {
	return func(c *Options) { c.WatchedYear = year }
}

// WithGenre will filter the diary and list by a genre
func WithGenre(genre string) Option {
	return func(c *Options) { c.Genre = strings.ToLower(genre) }
}

// WithDecade will filter the diary and list by decade of movie release
func WithDecade(decade string) Option {
	return func(c *Options) { c.Decade = decade }
}

// WithYear will filter the diary and list by year of movie release
func WithYear(year string) Option {
	return func(c *Options) { c.Year = year }
}

// WithDirector will filter the diary and list by director slug
func WithDirector(director string) Option {
	return func(c *Options) { c.Director = director }
}

// WithActor will filter the diary and list by actor slug
func WithActor(actor string) Option {
	return func(c *Options) { c.Actor = actor }
}

// WithRating will filter the diary by rating
func WithRating(rating Rating) Option {
	return func(c *Options) { c.Rating = rating }
}

// build will build the url path with the filters and sortby options
func (f Options) build() string {
	segments := make([]string, 0)
	if f.WatchedYear != "" {
		segments = append(segments, "for", f.WatchedYear)
	}
	if f.Genre != "" {
		segments = append(segments, "genre", f.Genre)
	}
	if f.Decade != "" {
		segments = append(segments, "decade", f.Decade)
	}
	if f.Year != "" {
		segments = append(segments, "year", f.Year)
	}
	if f.Director != "" {
		segments = append(segments, "with/director", f.Director)
	}
	if f.Actor != "" {
		segments = append(segments, "with/actor", f.Actor)
	}
	if f.Rating != "" {
		segments = append(segments, "rated", f.Rating.String())
	}
	if f.Sort != "" {
		segments = append(segments, f.Sort.String())
	}
	return strings.Join(segments, "/")
}
