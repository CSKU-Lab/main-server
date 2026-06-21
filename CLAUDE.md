## Wire Dependency Injection

**NEVER manually edit `cmd/app/wire_gen.go`.**

`wire_gen.go` is auto-generated. After adding providers or changing constructor signatures:

```bash
cd cmd/app && go run -mod=mod github.com/google/wire/cmd/wire
```

Add new providers to the appropriate set in `internal/providers/repositories.go` or `internal/providers/services.go`. Wire resolves injection automatically from those sets.
