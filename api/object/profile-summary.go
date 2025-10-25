package object

type HeapProfileSummary struct {
	TimeNanos    int64 `json:"time_nanos"`    // Time of collection (UTC) represented as nanoseconds past the epoch
	InuseSpace   int64 `json:"inuse_space"`   // Total bytes currently allocated
	InuseObjects int64 `json:"inuse_objects"` // Total objects currently allocated
	AllocSpace   int64 `json:"alloc_space"`   // Total bytes allocated (including freed)
	AllocObjects int64 `json:"alloc_objects"` // Total objects allocated (including freed)
}
