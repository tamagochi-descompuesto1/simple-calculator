// Types for the environment variables this app reads through import.meta.env.
// Vite only exposes variables whose name starts with VITE_, and it embeds their
// values at build time. See .env.example.
interface ImportMetaEnv {
    readonly VITE_API_URL?: string;
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}
