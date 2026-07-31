/** @type {import('tailwindcss').Config} */
// Prospecção Brasil -- Real Estate Intelligence design system.
// Canonical token set from the brand spec. The default Tailwind palette is
// not used in UI; only the tokens below are allowed. Any color, size,
// spacing, or radius not defined here is a bug.
module.exports = {
  content: [
    "./internal/ui/templates/**/*.html",
    "./internal/handler/templates/**/*.html",
    "./**/*.go",
  ],
  theme: {
    extend: {
      colors: {
        // Surface scale (warm off-white, the "clean" premium background)
        surface: {
          DEFAULT: "#fcf9f8",
          dim: "#dcd9d9",
          bright: "#fcf9f8",
          "container-lowest": "#ffffff",
          "container-low": "#f6f3f2",
          container: "#f0edec",
          "container-high": "#ebe7e7",
          "container-highest": "#e5e2e1",
        },
        "on-surface": "#1c1b1b",
        "on-surface-variant": "#44474e",
        "inverse-surface": "#313030",
        "inverse-on-surface": "#f3f0ef",
        outline: "#75777f",
        "outline-variant": "#c5c6cf",
        "surface-tint": "#4e5e82",
        "surface-variant": "#e5e2e1",

        // Primary: Deep Navy -- "Corporate Intelligence" weight
        primary: {
          DEFAULT: "#031636",
          container: "#1a2b4c",
          "on-container": "#8293ba",
          fixed: "#d8e2ff",
          "fixed-dim": "#b6c6f0",
          "on-fixed": "#071b3b",
          "on-fixed-variant": "#364669",
        },
        "on-primary": "#ffffff",
        "inverse-primary": "#b6c6f0",

        // Secondary: Sóbrio Gold -- high-value CTAs and key metrics
        secondary: {
          DEFAULT: "#765a1a",
          container: "#ffd88b",
          "on-container": "#795c1c",
          fixed: "#ffdea1",
          "fixed-dim": "#e7c177",
          "on-fixed": "#261900",
          "on-fixed-variant": "#5c4301",
        },
        "on-secondary": "#ffffff",

        // Tertiary: warm dark brown
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
        background: "#fcf9f8",
        "on-background": "#1c1b1b",
        "background-alt": "#F8FAFC",
        "slate-gray": "#334155",
        "whatsapp-green": "#25D366",
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
        // Display
        "display-lg": ["48px", { lineHeight: "1.1", fontWeight: "700", letterSpacing: "-0.02em" }],
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
        // Ambient shadows: extremely soft, large-radius, low-opacity Navy
        // (Blur: 20px, Opacity: 4%, Color: #1A2B4C) -- "floated" but stable
        DEFAULT: "0 4px 20px rgba(26, 43, 76, 0.04)",
        sm: "0 2px 10px rgba(26, 43, 76, 0.03)",
        md: "0 4px 20px rgba(26, 43, 76, 0.04)",
        lg: "0 8px 30px rgba(26, 43, 76, 0.05)",
        xl: "0 12px 40px rgba(26, 43, 76, 0.06)",
        focus: "0 0 0 2px #ffffff, 0 0 0 4px #031636",
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
