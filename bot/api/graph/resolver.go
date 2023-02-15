package graph

import "github.com/tashima42/awa-bot/bot/pkg/db"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Repo *db.Repo
}
