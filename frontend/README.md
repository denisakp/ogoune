# Pulseguard Frontend

A modern and responsive dashboard for the Pulseguard monitoring platform, built with Vue 3 and Ant Design.

## Technology Stack

-   **Framework**: [Vue 3](https://vuejs.org/) (using Composition API and `<script setup>`)
-   **Language**: [TypeScript](https://www.typescriptlang.org/)
-   **UI Library**: [Ant Design Vue](https://www.antdv.com/)
-   **Build Tool**: [Vite](https://vitejs.dev/)
-   **Routing**: [Vue Router](https://router.vuejs.org/)
-   **HTTP Client**: [Axios](https://axios-http.com/)

## Project Structure

The `src` directory is organized by feature, separating concerns into distinct layers.

```
src/
├── assets/         # Static assets (images, fonts)
├── components/     # Reusable UI components (buttons, forms, etc.)
├── composables/    # State management and business logic
├── libs/           # Third-party library configurations (e.g., Axios)
├── router/         # Application routing configuration
├── services/       # API communication layer
├── stores/         # Pinia stores for global state
├── types/          # TypeScript interfaces and type definitions
├── views/          # Page-level components
├── App.vue         # Root application component and layout
└── main.ts         # Application entry point
```

For a detailed explanation of the architecture, data flow, and coding patterns, please see [ARCHITECTURE.md](ARCHITECTURE.md).

## Getting Started

### Prerequisites

-   [Node.js](https://nodejs.org/) (version 20.x or higher)
-   [pnpm](https://pnpm.io/)

### 1. Install Dependencies

Navigate to the `frontend` directory and install the required packages.

```bash
cd frontend
pnpm install
```

### 2. Configure Environment Variables

Create a `.env.local` file in the `frontend` directory. This file will store the URL of your backend API.

```env
# .env.local
VITE_API_BASE_URL=http://localhost:8080
```

### 3. Run the Development Server

Start the Vite development server. Make sure your backend server is running, as the frontend will make requests to it.

```bash
pnpm dev
```

The application will be available at `http://localhost:5173`.

## Available Scripts

-   `pnpm dev`: Starts the development server with hot-reloading.
-   `pnpm build`: Compiles the application for production.
-   `pnpm preview`: Serves the production build locally for testing.
-   `pnpm lint`: Lints the codebase to find and fix problems.
-   `pnpm format`: Formats all files with Prettier.
-   `pnpm type-check`: Runs the TypeScript compiler to check for type errors.

## License

This project is licensed under the MIT License.