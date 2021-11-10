package models

type Entry struct {
	Offset    int       `json:"offset"`
	Operation Operation `json:",inline"`
}
