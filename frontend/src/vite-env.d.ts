// Types for the VITE_ variables read through import.meta.env. Vite embeds them
// at build time. See .env.example.
interface ImportMetaEnv {
  readonly VITE_API_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
