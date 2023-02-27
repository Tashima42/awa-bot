package auth

import (
	"context"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"log"
	"net/http"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func Middleware(repo *db.Repo) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var apikey string
			apikey = r.Header.Get("Authorization")
			log.Println("apikey: " + apikey)
			if apikey == "" {
				cookie, err := r.Cookie("apikey")
				if err != nil {
					http.Error(w, "Unauthorized: Missing apikey", http.StatusUnauthorized)
					return
				}
				apikey = cookie.Value
			}
			user, err := repo.GetUserByApiKey(r.Context(), apikey)
			if err != nil {
				log.Println(err)
				http.Error(w, "Unauthorized: Invalid apikey", http.StatusUnauthorized)
				return
			}
			// add user to context
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			// set new context with user
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) *db.User {
	raw, _ := ctx.Value(userCtxKey).(*db.User)
	return raw
}
