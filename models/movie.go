package models

type Movie struct {
	ID          int
	CategoryID 	int
	Name 		string
	IsVip		bool
	FullUrl 	string
	CoverImageUrl 		string
	Tags 		string
	TsDownloaded bool
}