package formatter

type Status string

var (
	Success             Status = "00"
	CacheError          Status = "PAKU01"
	DatabaseError       Status = "PAKU02"
	InvalidRequest      Status = "PAKU03"
	DataNotFound        Status = "PAKU04"
	InternalServerError Status = "PAKU05"
	DataConflict        Status = "PAKU06"
	Unauthorized        Status = "PAKU07"
)

func (s Status) String() string {
	return string(s)
}
