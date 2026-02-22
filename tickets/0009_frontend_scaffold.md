# 0009 — Frontend Scaffold

## Goal

Initialise the Svelte 5 + Vite + Tailwind CSS project in `frontend/`, confirm it builds, and embed the output into the Go binary so it is served at `/`.

## Dependencies

- Ticket 0001 (Go module)
- Ticket 0006 (router — has the `// TODO: embed SPA` placeholder)

## Commands

```sh
# Scaffold Svelte + Vite project (select Svelte, TypeScript)
npm create vite@latest frontend -- --template svelte-ts
cd frontend
npm install

# Add Tailwind CSS v4
npm install tailwindcss @tailwindcss/vite

# Add svelte-dnd-action
npm install svelte-dnd-action
```

Configure Tailwind in `frontend/vite.config.ts`:
```ts
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [tailwindcss(), svelte()],
})
```

Add to the top of `frontend/src/app.css`:
```css
@import "tailwindcss";
```

Build:
```sh
cd frontend && npm run build
```

## Embedding

Create `internal/api/embed.go`:

```go
package api

import "embed"

//go:embed all:../../frontend/dist
var frontendFS embed.FS
```

In `internal/api/router.go`, replace the `// TODO: embed SPA` comment:

```go
// Serve embedded SPA for all non-API routes
frontendDist, _ := fs.Sub(frontendFS, "frontend/dist")
fileServer := http.FileServer(http.FS(frontendDist))
r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
    // Serve index.html for all unmatched routes (SPA client-side routing)
    _, err := frontendDist.(fs.ReadFileFS).ReadFile(r.URL.Path[1:])
    if err != nil {
        r.URL.Path = "/"
    }
    fileServer.ServeHTTP(w, r)
})
```

## Acceptance Criteria

- `cd frontend && npm run build` exits 0 and produces `frontend/dist/`.
- `go build ./...` exits 0 (embed compiles successfully).
- Running the binary and opening `http://localhost:8080` serves the Vite/Svelte default page.
- The API is still reachable at `http://localhost:8080/api/v1/boards`.
- `frontend/dist/` is listed in `.gitignore`.
