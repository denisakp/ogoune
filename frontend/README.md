# Pulseguard Frontend

A modern, responsive Vue 3 + TypeScript + Daisy UI dashboard for the Pulseguard monitoring platform.

## Technology Stack

- **Framework**: Vue 3 (Composition API with `<script setup>`)
- **Language**: TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS + Daisy UI
- **HTTP Client**: Axios
- **Routing**: Vue Router v4
- **State Management**: Pinia (for future use)

## Project Structure

```
src/
├── components/          # Reusable Vue components
│   ├── ErrorAlert.vue
│   ├── LoadingSpinner.vue
│   ├── ResourceForm.vue
│   └── StatusBadge.vue
├── composables/         # Vue composable functions for state management
│   ├── useActivities.ts
│   ├── useIntegrations.ts
│   ├── useResources.ts
│   └── useTags.ts
├── services/            # API client and service layer
│   ├── apiClient.ts
│   ├── activityService.ts
│   ├── integrationService.ts
│   ├── resourceService.ts
│   └── tagService.ts
├── types/               # TypeScript interfaces and types
│   └── index.ts
├── views/               # Page components
│   ├── ActivitiesView.vue
│   ├── IntegrationsView.vue
│   ├── MonitorsView.vue
│   └── TagsView.vue
├── router/              # Vue Router configuration
│   └── index.ts
├── App.vue              # Root component with navigation
├── main.ts              # Application entry point
└── style.css            # Global styles
```

## Architectural Principles

### 1. Separation of Concerns

- **Services** (`src/services/`): All axios/HTTP logic is abstracted here. Components never call APIs directly.
- **Composables** (`src/composables/`): State management (loading, error, data) for pages.
- **Components** (`src/components/`): UI rendering only, delegating logic to composables.
- **Views** (`src/views/`): Page-level components using composables.

### 2. Type Safety

All API responses, component props, and state are strongly typed using TypeScript interfaces in `src/types/index.ts`.

### 3. Composition API

All components use `<script setup lang="ts">` syntax for cleaner, more readable code.

## Getting Started

### Prerequisites

- Node.js 20.19.0 or >=22.12.0
- pnpm (or npm)

### Installation

```bash
cd frontend
pnpm install
```

### Configuration

Create a `.env.local` file for development:

```env
VITE_API_BASE_URL=http://localhost:8080
```

For production, use `.env.production`:

```env
VITE_API_BASE_URL=/api
```

### Development Server

```bash
pnpm run dev
```

The application will be available at `http://localhost:5173` (or the next available port).

### Build for Production

```bash
pnpm run build
```

The production build will be in the `dist/` directory.

### Lint and Format

```bash
pnpm run lint       # Run ESLint and Oxlint
pnpm run format     # Format with Prettier
```

## API Integration

The frontend communicates with the backend API at `http://localhost:8080` (configurable via environment variables).

### Available Endpoints

- **Monitors**: GET, POST, PATCH, DELETE `/resources`
- **Tags**: GET, POST, PATCH, DELETE `/tags`
- **Integrations**: GET, POST, PATCH `/integrations`
- **Activities**: GET `/monitoring-activities`

See `src/services/` for service implementations.

## Features

### Monitors (Resources)

- View all monitors in a table with status, target, and last check time
- Create new monitors with HTTP/TCP types
- Edit existing monitors
- Pause/Resume monitoring
- Delete monitors
- Real-time status badges

### Tags

- Organize monitors with tags
- Create, edit, and delete tags
- Associate tags with resources (future enhancement)

### Integrations

- Configure notification channels (SMTP, Slack, Discord, Google Chat, Webhook)
- Enable/disable integrations
- Filter notifications by event type (Up/Down)

### Activities

- View all monitoring check results
- See response times and success/failure status
- Filter by resource (future enhancement)
- Real-time updates via WebSocket (future enhancement)

## Future Enhancements

- [ ] WebSocket integration for real-time activity updates
- [ ] Resource filtering and search
- [ ] Advanced dashboard with analytics
- [ ] User authentication and authorization
- [ ] Dark/light theme toggle
- [ ] Incident timeline and history
- [ ] Notification preview and test
- [ ] Bulk operations on resources

## Styling

The application uses Daisy UI components with a dark theme (`data-theme="night"`) configured in `tailwind.config.ts`.

All interactive elements use Daisy UI classes:

- Buttons: `btn`, `btn-primary`, `btn-error`, etc.
- Tables: `table`, `table-compact`
- Modals: `modal`, `modal-box`
- Forms: `form-control`, `input`, `select`, `checkbox`, etc.

## Error Handling

- All async operations are wrapped with try-catch
- Errors are displayed using the `ErrorAlert` component
- Loading states are managed with `loading` refs in composables
- Network errors are logged to the browser console

## Performance Considerations

- Lazy loading with Vue Router (can be implemented as routes grow)
- Memoization with composables (data is cached and reused)
- Efficient re-renders with `<script setup>` and reactive APIs

## Contributing

When adding new features:

1. Create service functions in `src/services/` if API calls are needed
2. Create composables in `src/composables/` for state management
3. Create reusable components in `src/components/`
4. Create views in `src/views/` for new pages
5. Add types to `src/types/index.ts`
6. Add routes to `src/router/index.ts`

Maintain the separation of concerns: no components should call APIs directly.

## License

MIT - see LICENSE file in the repository root.
