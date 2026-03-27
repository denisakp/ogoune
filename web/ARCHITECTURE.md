# Frontend Architecture Guide

This document provides a comprehensive overview of the Ogoune frontend architecture. It is designed to help developers understand the project structure, data flow, and core principles, enabling them to contribute effectively and maintain code quality.

## 1. Core Principles & Data Flow

The architecture is built on a strict separation of concerns, ensuring that UI, state management, and API communication are decoupled. This makes the codebase easier to understand, test, and scale.

### Data Flow

The application follows a unidirectional data flow pattern:

-   **Request Flow**: A user interaction in a `Component` triggers a function in a `Composable`, which in turn calls a `Service` to make an API request.
-   **Response Flow**: The `Service` receives a response from the API, the `Composable` updates its reactive state with the new data, and the `Component` automatically re-renders to reflect the changes.

```
Request:  Component → Composable → Service → API Client → Backend
Response: Backend → API Client → Service → Composable (updates state) → Component (re-renders)
```

### Guiding Rule: No Direct API Calls from Components

Components should never contain direct HTTP requests (e.g., using `axios`). All API interactions must be delegated to the service layer and managed by composables.

-   **Incorrect Practice**:
    ```vue
    <script setup>
    import axios from 'axios'
    onMounted(async () => {
      const res = await axios.get('/resources/123')
      // ...
    })
    </script>
    ```

-   **Correct Practice**:
    ```vue
    <script setup>
    import { useResources } from '@/composables/useResources'
    const { resources, loadResources } = useResources()
    onMounted(() => loadResources())
    </script>
    ```

## 2. File Organization

The project follows a feature-oriented structure within the `src/` directory.

```
src/
├── assets/               # Static assets (images, fonts)
├── components/           # Reusable, stateless UI components (e.g., buttons, badges)
├── composables/          # State management and business logic (e.g., useResources.ts)
├── libs/                 # Third-party library configurations (e.g., axios.helper.ts)
├── router/               # Vue Router configuration (index.ts)
├── services/             # API communication layer (e.g., resourceService.ts)
├── stores/               # Pinia stores for global state (if needed)
├── types/                # TypeScript interfaces and type definitions (index.ts)
├── views/                # Page-level components, mapped to routes
├── App.vue               # Root application component with layout and navigation
├── main.ts               # Application entry point
└── style.css             # Global CSS styles
```

## 3. Architectural Layers

### 3.1. Views (`src/views/`)

-   **Purpose**: Page-level components that correspond to a specific route (e.g., `MonitorsView.vue`).
-   **Responsibilities**:
    -   Compose the page layout using smaller, reusable components from `src/components/`.
    -   Utilize one or more composables from `src/composables/` to fetch and manage the data required for the page.
    -   Handle user events by calling methods exposed by the composables.
    -   Display loading, error, and data states.

### 3.2. Components (`src/components/`)

-   **Purpose**: Small, reusable UI elements (e.g., `StatusBadge.vue`, `ResourceForm.vue`).
-   **Responsibilities**:
    -   Receive data via `props`.
    -   Emit events to parent components via `emits`.
    -   Contain minimal to no business logic.
    -   Be as stateless as possible.

### 3.3. Composables (`src/composables/`)

-   **Purpose**: The core of the state management and business logic. Each composable typically manages a specific piece of domain state (e.g., `useResources` manages the list of monitoring resources).
-   **Responsibilities**:
    -   Define reactive state variables (`ref`, `reactive`) for data, loading status, and errors.
    -   Expose functions that orchestrate calls to the `Service` layer to perform CRUD operations.
    -   Handle the `try/catch/finally` logic for asynchronous operations, updating loading and error states accordingly.
    -   Return the reactive state and methods for the `View` to consume.

-   **Example (`useResources.ts`)**:
    ```typescript
    import { ref } from 'vue'
    import * as resourceService from '@/services/resourceService'
    import type { Resource } from '@/types'

    export function useResources() {
      const resources = ref<Resource[]>([])
      const loading = ref(false)
      const error = ref<string | null>(null)

      const loadResources = async () => {
        loading.value = true
        error.value = null
        try {
          resources.value = await resourceService.fetchResources()
        } catch (err) {
          error.value = err instanceof Error ? err.message : 'Failed to load resources'
        } finally {
          loading.value = false
        }
      }

      // ... other methods like createResource, deleteResource

      return { resources, loading, error, loadResources }
    }
    ```

### 3.4. Services (`src/services/`)

-   **Purpose**: The sole layer responsible for communicating with the backend API.
-   **Responsibilities**:
    -   Export async functions that correspond to specific API endpoints (e.g., `fetchResources`, `createResource`).
    -   Use the centralized `apiClient` to make HTTP requests.
    -   Handle data transformation if the API response needs to be adapted for the frontend.
    -   Define the expected data types for requests and responses, using types from `src/types/index.ts`.

-   **Example (`resourceService.ts`)**:
    ```typescript
    import apiClient from '@/libs/axios.helper'
    import type { Resource } from '@/types'

    export const fetchResources = async (): Promise<Resource[]> => {
      const response = await apiClient.get('/resources')
      return response.data.data || []
    }

    export const createResource = async (payload: Omit<Resource, 'id'>): Promise<Resource> => {
      const response = await apiClient.post('/resources', payload)
      return response.data.data
    }
    ```

### 3.5. API Client (`src/libs/axios.helper.ts`)

-   **Purpose**: A single, pre-configured Axios instance used by all services.
-   **Responsibilities**:
    -   Set the `baseURL` from the `VITE_API_BASE_URL` environment variable.
    -   Configure default headers (e.g., `Content-Type: application/json`).
    -   Set up interceptors for centralized request/response handling (e.g., logging, error transformation, adding auth tokens).

## 4. Error Handling

Error handling is managed consistently across layers:

1.  **Service Layer**: API calls made with `apiClient` will throw an error on non-2xx responses. Services do not `try/catch` these errors; they let them propagate up.
2.  **Composable Layer**: This is where errors are caught. Every async function that calls a service is wrapped in a `try/catch` block.
    -   On failure, the `error` ref is populated with a user-friendly message.
    -   The `loading` ref is always set to `false` in a `finally` block to ensure the UI is never stuck in a loading state.
3.  **View/Component Layer**: The UI uses a `v-if="error"` directive to conditionally render an error message or an alert component, displaying the content of the `error` ref.

## 5. Type Safety

TypeScript is used throughout the project to ensure type safety and improve developer experience.

-   **`src/types/index.ts`**: This file is the single source of truth for all data structures (e.g., `Resource`, `Tag`, `Integration`).
-   **Flow of Types**:
    1.  Service functions use these types for their arguments and return values.
    2.  Composables use them to type their reactive state (`ref<Resource[]>([])`).
    3.  Components and Views use them for `props` and to correctly access data properties.

This end-to-end typing prevents common bugs and provides excellent autocompletion in the IDE.

## 6. Development Workflow: Adding a New Feature

Follow these steps to add a new page or feature (e.g., a "Status Page"):

1.  **Define Types** (if new data structures are needed) in `src/types/index.ts`.
2.  **Create Service Function(s)** in a new or existing file in `src/services/` to fetch data from the API.
3.  **Create a Composable** in `src/composables/` to manage the state and logic for the feature.
4.  **Create a View** component in `src/views/` that uses the composable and renders the UI.
5.  **Add a Route** in `src/router/index.ts` to map a URL path to your new view.
6.  **Update Navigation** (e.g., in `App.vue`) to link to the new page.