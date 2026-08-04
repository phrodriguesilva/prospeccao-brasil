/** @type {import('tailwindcss').Config} */
// Prospecção Brasil -- Dark Premium design system.
// Black / charcoal surfaces with metallic gold accents. The default Tailwind
// palette is not used in UI; only the tokens below are allowed. Any color,
// size, spacing, or radius not defined here is a bug.
module.exports = {
  content: [
    "./internal/template/**/*.html",
    "./**/*.go",
  ],
  theme: {
    extend: {
      colors: {
        // Surface scale (deep black / charcoal -- the premium dark background)
        surface: {
          DEFAULT: "#121212",
          dim: "#0a0a0a",
          bright: "#161616",
          "container-lowest": "#0e0e0e",
          "container-low": "#161616",
          container: "#1c1c1c",
          "container-high": "#222222",
          "container-highest": "#282828",
        },
        "on-surface": "#ffffff",
        "on-surface-variant": "#a0a0a0",
        "inverse-surface": "#f3f0ef",
        "inverse-on-surface": "#1c1b1b",
        outline: "#4a4a4a",
        "outline-variant": "#2e2e2e",
        "surface-tint": "#c8a25d",
        "surface-variant": "#242424",

        // Primary: Metallic Gold -- the high-value accent (CTAs, icons, lines)
        primary: {
          DEFAULT: "#c8a25d",
          container: "#d4af6a",
          "on-container": "#121212",
          fixed: "#d4af6a",
          "fixed-dim": "#c8a25d",
          "on-fixed": "#121212",
          "on-fixed-variant": "#7a5e2a",
        },
        "on-primary": "#121212",
        "inverse-primary": "#1a2b4c",

        // Brand aliases
        "navy-brand": "#031636",
        "gold-brand": "#c8a25d",
        "gold-light": "#d4af6a",
        "gold-lightest": "#f5e6c8",
        "gold-dark": "#765a1a",
        "cream-brand": "#fcf9f8",

        // Secondary: Sóbrio Gold ramp -- gradients and subtle highlights
        secondary: {
          DEFAULT: "#765a1a",
          light: "#d4af6a",
          lightest: "#f5e6c8",
          container: "#d4af6a",
          "on-container": "#121212",
          fixed: "#d4af6a",
          "fixed-dim": "#c8a25d",
          "on-fixed": "#121212",
          "on-fixed-variant": "#7a5e2a",
        },
        "on-secondary": "#121212",

        // Tertiary: warm dark brown (deep accents)
        tertiary: {
          DEFAULT: "#241300",
          container: "#3f2600",
          "on-container": "#b28c5b",
          fixed: "#ffddb5",
          "fixed-dim": "#eabf8a",
          "on-fixed": "#2a1800",
          "on-fixed-variant": "#5e4117",
        },
        "on-tertiary": "#ffffff",

        // Error
        error: {
          DEFAULT: "#ba1a1a",
          container: "#ffdad6",
          "on-container": "#93000a",
        },
        "on-error": "#ffffff",

        // Semantic aliases
        background: "#121212",
        "on-background": "#ffffff",
        "background-alt": "#0a0a0a",
        "slate-gray": "#a0a0a0",
        "whatsapp-green": "#25D366",

        // Neutral aliases for admin data tables (dark-on-dark borders)
        neutral: {
          border: "#2e2e2e",
          "border-hover": "#4a4a4a",
          muted: "#6a6a6a",
          text: "#ffffff",
        },
      },
      fontFamily: {
        sans: [
          "Inter",
          "system-ui",
          "-apple-system",
          "Segoe UI",
          "Roboto",
          "Helvetica Neue",
          "Arial",
          "sans-serif",
        ],
        display: [
          "Montserrat",
          "Inter",
          "system-ui",
          "-apple-system",
          "Segoe UI",
          "sans-serif",
        ],
        mono: ["ui-monospace", "SFMono-Regular", "Menlo", "Consolas", "monospace"],
      },
      fontSize: {
        // Display -- responsive via clamp (hero headlines)
        "display-lg": ["clamp(2.5rem, 5vw, 4.5rem)", { lineHeight: "1.1", fontWeight: "800", letterSpacing: "-0.02em" }],
        // Headlines (Montserrat)
        "headline-lg": ["32px", { lineHeight: "1.2", fontWeight: "600" }],
        "headline-lg-mobile": ["28px", { lineHeight: "1.2", fontWeight: "600" }],
        "headline-md": ["24px", { lineHeight: "1.3", fontWeight: "600" }],
        // Body (Inter)
        "body-lg": ["18px", { lineHeight: "1.6", fontWeight: "400" }],
        "body-md": ["16px", { lineHeight: "1.6", fontWeight: "400" }],
        // Label (Inter, uppercase, tracked)
        "label-sm": ["12px", { lineHeight: "1", fontWeight: "600", letterSpacing: "0.05em" }],
      },
      borderRadius: {
        sm: "0.125rem",
        DEFAULT: "0.25rem",
        md: "0.375rem",
        lg: "0.5rem",
        xl: "0.75rem",
        full: "9999px",
      },
      boxShadow: {
        // Ambient shadows: deep black, low-opacity -- "floated" on dark
        DEFAULT: "0 4px 20px rgba(0, 0, 0, 0.5)",
        sm: "0 2px 10px rgba(0, 0, 0, 0.4)",
        md: "0 4px 20px rgba(0, 0, 0, 0.5)",
        lg: "0 8px 30px rgba(0, 0, 0, 0.6)",
        xl: "0 12px 40px rgba(0, 0, 0, 0.7)",
        focus: "0 0 0 2px #121212, 0 0 0 4px #c8a25d",
      },
      spacing: {
        "section-gap": "80px",
        "margin-mobile": "20px",
        "gutter-mobile": "16px",
        "stack-sm": "8px",
        "stack-md": "16px",
        "stack-lg": "32px",
      },
      width: {
        topbar: "100%",
      },
      zIndex: {
        sticky: "10",
        topbar: "30",
        modal: "40",
      },
    },
  },
  plugins: [],
};
