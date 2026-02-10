package models

type Product struct {
	ProductID   string   `json:"product_id"`
	CategoryID  string   `json:"category_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	ImageURL    string   `json:"image_url"`
	Photos      []string `json:"photos"`
	Quantity    int64    `json:"quantity"`
	Rating      int64    `json:"rating"`
}
