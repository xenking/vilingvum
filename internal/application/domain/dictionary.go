package domain

type Dictionary struct {
	Data  []DictRecord
	Index []int
}

type DictRecord struct {
	Word    string
	Meaning string
	TopicID int64
	ID      int64
}
