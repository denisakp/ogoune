# Pulseguard Frontend

A modern, responsive dashboard for the Pulseguard monitoring platform, built with Vue 3 and TypeScript.

---

## 🚀 Quick Start

### Prerequisites

- Node.js 22+
- pnpm (recommended) or npm
- Backend API running at `http://localhost:8080/api` (for local development)

### 1. Install Dependencies

```bash
cd frontend
pnpm install
```

### 2. Configure Environment

Create a `.env.local` file:

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:8080/api
```

**Note:** This is only needed for local development. When deployed with the backend, requests automatically go to `/api` on the same origin.

### 3. Start Development Server

```bash
pnpm dev
```

Dashboard available at: `http://localhost:5173`

---

## 📦 Available Scripts

| Command | Purpose |
|---------|---------|
| `pnpm install` | Install dependencies |
| `pnpm dev` | Start development server with hot reload |
| `pnpm build` | Build for production (generates `dist/` folder) |
| `pnpm preview` | Preview production build locally |
| `pnpm lint` | Lint code for issues |
| `pnpm format` | Format code with Prettier |
| `pnpm type-check` | Check TypeScript types without building |

---

## 🎯 Features

### Monitor Management
- **Create** – Add HTTP/HTTPS or TCP monitors
- **View** – Real-time status and uptime metrics
- **Edit** – Update monitor configuration (intervals, timeouts, etc.)
- **Delete** – Remove monitors safely
- **Pause/Resume** – Temporarily stop monitoring without deletion
- **Organize** – Tag monitors for better organization

### Incident Tracking
- **View** – Complete incident history
- **Timeline** – Event step tracking (detected, resolved, etc.)
- **Filter** – Show only unresolved incidents
- **Duration** – Track incident length and impact

### Activity Monitoring
- **Log** – View all health check activities
- **Filter** – By resource or status
- **Response Times** – Track performance metrics
- **Errors** – See detailed failure reasons

### Status Page
- **Public Access** – Share read-only status with stakeholders
- **Uptime Display** – 90-day uptime trends and metrics
- **Incident History** – Full list of incidents
- **Real-time Updates** – Status reflects changes immediately

### Statistics
- **Global Uptime** – System-wide uptime percentage
- **Incident Count** – Total and unresolved incidents
- **Time Ranges** – View stats for 24h, 7d, 30d, or 90d

---

## ⚙️ Configuration

### Environment Variables

#### Development

Create `.env.local`:

```bash
# Backend API URL (optional; defaults to /api)
VITE_API_BASE_URL=http://localhost:8080/api
```

#### Production

When deployed with the backend:
- No environment variables needed
- Frontend requests go to `/api/...` on the same origin
- Backend serves both static files and API

---

## 📚 Development

### Project Structure

```
frontend/src/
├── assets/                 Static assets (images, fonts, styles)
├── components/             Reusable UI components
├── composables/            State and business logic (e.g., useResources.ts)
├── libs/
│   └── axios.helper.ts     Axios configuration with base URL
├── router/
│   └── index.ts            Vue Router setup
├── services/               API communication layer
│   ├── resourceService.ts
│   ├── incidentService.ts
│   └── ...
├── stores/                 Pinia global state
├── types/
│   └── index.ts            TypeScript type definitions
├── views/                  Page-level components
├── App.vue                 Root component with layout
├── main.ts                 Application entry point
└── style.css               Global styles
```

### Architecture & Patterns

The frontend follows a strict separation of concerns:

**Components** → **Composables** → **Services** → **API**

- **Components** are presentational (receive props, emit events)
- **Composables** manage state and orchestrate business logic
- **Services** handle all HTTP communication
- Never call Axios directly from components

For detailed architectural information, see [Frontend Architecture](./ARCHITECTURE.md).

### Adding a New Feature

1. **Define types** in `src/types/index.ts`
2. **Create service** in `src/services/` for API calls
3. **Create composable** in `src/composables/` for state management
4. **Create view** in `src/views/` as the page component
5. **Add route** in `src/router/index.ts`
6. **Update navigation** in `src/App.vue`

---

## 🧪 Testing

```bash
# Type checking
pnpm type-check

# Linting
pnpm lint

# Format
pnpm format
```

---

## 🏗️ Building for Production

### Build Step

```bash
pnpm build
```

This generates a production-optimized build in the `dist/` folder.

### Deployment Options

**With Backend:**
1. Build the frontend: `pnpm build`
2. Copy `dist/` contents to backend's static directory
3. Backend serves both API and frontend at one origin

**Standalone (SPA):**
1. Build the frontend: `pnpm build`
2. Deploy `dist/` to a CDN or web server
3. Configure API base URL via reverse proxy or environment

---

## 🐛 Troubleshooting

### API Requests Return 404

**Problem:** Requests fail with 404 Not Found

**Solutions:**
- Verify backend is running: `curl http://localhost:8080/health`
- Verify `VITE_API_BASE_URL` in `.env.local` is correct
- Check API endpoint paths (e.g., `/api/resources` not `/resources`)
- Ensure backend URL includes `/api/` path

### Components Not Updating

**Problem:** State changes don't trigger re-renders

**Solutions:**
- Ensure store state is used in `<template>` or `computed()`
- Don't mutate store state directly; use store actions
- Verify composable is imported correctly
- Check that reactive refs are properly initialized

### TypeScript Errors in IDE

**Problem:** Type errors even though code works

**Solutions:**
- Run `pnpm type-check` to see actual errors
- Verify types in `src/types/index.ts` match backend responses
- Ensure all API response types are properly defined
- Update `tsconfig.json` if needed

### Port 5173 Already in Use

**Problem:** `EADDRINUSE: address already in use :::5173`

**Solutions:**
- Stop the process using port 5173
- Or specify a different port: `pnpm dev -- --port 5174`

### Blank Page or CORS Errors

**Problem:** Frontend loads but shows blank or CORS error in console

**Solutions:**
- Verify backend is running and accessible
- Check `VITE_API_BASE_URL` points to correct backend URL
- Ensure backend has CORS enabled for your frontend URL
- Check browser console for detailed error messages

---

## 🎨 UI Components

The frontend uses [Ant Design Vue](https://www.antdv.com/) components:

- `AButton` – Clickable buttons
- `ATable` – Data tables with sorting/pagination
- `AForm` – Forms with validation
- `AModal` – Dialog modals
- `ACard` – Content containers
- `ATag` – Labels and tags
- `ASpin` – Loading spinners
- `AMessage` / `ANotification` – Alerts and toasts

See [Ant Design Vue documentation](https://www.antdv.com/docs/vue/introduce) for all available components.

---

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)

---

## Performance

- **Code Splitting** – Vite automatically optimizes chunks
- **Lazy Loading** – Use `defineAsyncComponent` for heavy components
- **Virtual Scrolling** – For large data lists (Ant Design built-in)
- **Caching** – Axios can cache GET requests

---

## Development Workflow

### Local Setup
```bash
# Terminal 1: Start backend
cd backend && go run ./cmd/api

# Terminal 2: Start frontend
cd frontend && pnpm dev
```

### Making Changes
- Frontend hot-reloads on file changes
- TypeScript errors show in IDE and console
- Run `pnpm type-check` before committing

### Before Committing
```bash
pnpm lint      # Fix linting issues
pnpm format    # Format code
pnpm type-check # Check types
```

---

## Technical Details

For in-depth information on:

- **Architecture & Patterns** – See [Frontend Architecture](./ARCHITECTURE.md)
- **State Management** – See [Frontend Architecture - Composables](./ARCHITECTURE.md#33-composables-srccomposables)
- **API Integration** – See [Frontend Architecture - Services](./ARCHITECTURE.md#34-services-srcservices)
- **Type Safety** – See [Frontend Architecture - Type Safety](./ARCHITECTURE.md#5-type-safety)

---

## Contributing

When contributing to the frontend:

1. Follow Vue 3 Composition API patterns
2. Use TypeScript for type safety
3. Add types to `src/types/index.ts`
4. Keep components small and focused
5. Use services for all API calls
6. Run linting and formatting before committing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for more guidelines.

---

## 📄 License

MIT – see [LICENSE](../LICENSE)

---

## Related Documentation

- [Backend Setup](../backend/README.md) – Backend configuration and API
- [Root README](../README.md) – Project overview
- [Frontend Architecture](./ARCHITECTURE.md) – Technical design and patterns