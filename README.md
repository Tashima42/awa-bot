# awa-bot

## Dev
### Setup
* `cut -d' ' -f1 .tool-versions|xargs -I {} asdf plugin add {}`
* `asdf install`
* `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Migrations
* Create migration: `hack/create-migration migration_name`
* Migrate up: `hack/migrate up`
* Migrate down: `hack/migrate down`
