package store

const (
	FakeJSONArray = `[
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
		{"id": 3, "name": "Item 3"}
	]`

	FakeJSONObject = `{
		"items": [
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"}
		],
		"meta": {
			"total": 2,
			"page": 1
		}
	}`

	FakeNestedJSONArray = `[
		{
			"id": 1,
			"name": "Item 1",
			"tags": ["tag1", "tag2"]
		},
		{
			"id": 2,
			"name": "Item 2",
			"tags": ["tag3", "tag4"]
		}
	]`

	FakeNestedJSONObject = `{
		"users": [
			{
				"id": 1,
				"name": "User 1",
				"address": {
					"street": "123 Main St",
					"city": "Anytown"
				}
			},
			{
				"id": 2,
				"name": "User 2",
				"address": {
					"street": "456 Elm St",
					"city": "Othertown"
				}
			}
		]
	}`

	FakeNestedMixedData = `{
		"products": [
			{"id": 1, "name": "Product 1", "categories": ["cat1", "cat2"]},
			{"id": 2, "name": "Product 2", "categories": ["cat3", "cat4"]}
		],
		"orders": {
			"orderId": 123,
			"items": [
				{"productId": 1, "quantity": 2},
				{"productId": 2, "quantity": 1}
			]
		}
	}`
)
