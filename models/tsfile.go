package models

type Tsfile struct {
	ID          int
	MovieID 		int
	Xh 					int
	Name 				string
	Filesize 		int
	Time 				float32
	Finished 		bool
}

func (Tsfile) TableName() string{
	return "tsfiles"
}