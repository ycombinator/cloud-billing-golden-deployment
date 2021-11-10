package generator

type Config struct {
	StartOffsetSeconds int

	MaxCount         int
	MaxOffsetSeconds int

	MinIntervalSeconds int
	MaxIntervalSeconds int

	IndexToSearchRatio int
}
