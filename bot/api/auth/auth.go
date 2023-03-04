package auth

import (
	"context"
	"github.com/tashima42/awa-bot/bot/pkg/auth"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"log"
	"net/http"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func Middleware(repo *db.Repo, hashHelper *auth.HashHelper) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var apikey string
			var userID string
			apikey = r.Header.Get("Authorization")
			if apikey == "" {
				cookie, err := r.Cookie("apikey")
				if err != nil {
					http.Error(w, "Unauthorized: Missing apikey", http.StatusUnauthorized)
					return
				}
				apikey = cookie.Value
			}
			userID = r.Header.Get("X-UserID")
			if userID == "" {
				cookie, err := r.Cookie("userid")
				if err != nil {
					http.Error(w, "Unauthorized: Missing user id", http.StatusUnauthorized)
					return
				}
				userID = cookie.Value
			}
			userApiKey, err := repo.GetApiKeyByUserId(r.Context(), userID)
			if err != nil {
				log.Println(err)
				http.Error(w, "Unauthorized: Invalid apikey", http.StatusUnauthorized)
				return
			}
			if valid, err := hashHelper.Verify(apikey, userApiKey); err != nil || !valid {
				log.Println(err)
				http.Error(w, "Unauthorized: Invalid apikey", http.StatusUnauthorized)
				return
			}
			user, err := repo.GetUserByID(r.Context(), userID)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error: Failed to get user information", http.StatusInternalServerError)
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
