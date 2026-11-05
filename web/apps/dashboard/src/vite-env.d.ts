/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** The core URL. Empty: the same origin (the Vite proxy in development). */
  readonly VITE_API_BASE_URL?: string;
  /** Error tracking. Sentry is off when this is not set. */
  readonly VITE_SENTRY_DSN?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
