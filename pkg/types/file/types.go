package file

import "time"

// File struct
type File struct {
	Name         string
	Size         int64
	CreationTime time.Time
	ModifiedTime time.Time
	IsDir        bool
	FullPath     string
}
