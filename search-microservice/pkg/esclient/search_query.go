package esclient

type MultiMatch struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
}

type Bool struct {
	Must []any `json:"must"`
}

type Query struct {
	Bool Bool `json:"bool"`
}

type MultiMatchQuery struct {
	Query Query `json:"query"`
}

type ESHits[T any] struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source T `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
