package storage

import (
	pb_almanac "dinowernli.me/almanac/proto"
)

// OldestEntryFirst is an ordering over log entries by ascending timestamp.
type OldestEntryFirst []*pb_almanac.LogEntry

func (a OldestEntryFirst) Len() int           { return len(a) }
func (a OldestEntryFirst) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OldestEntryFirst) Less(i, j int) bool { return a[i].TimestampMs < a[j].TimestampMs }
