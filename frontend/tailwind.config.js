/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx,js,jsx}'],
  theme: {
    extend: {
      colors: {
        bg: '#0C0C0E',
        sidebar: '#09090C',
        surface: '#141417',
        elevated: '#1C1C20',
        border: 'rgba(255,255,255,0.06)',
        accent: '#6366F1',
        success: '#22D3A4',
        danger: '#F87171',
        warning: '#FBBF24',
      },
      fontFamily: {
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          "'SF Pro Text'",
          "'Segoe UI'",
          'system-ui',
          'sans-serif',
        ],
        mono: [
          "'JetBrains Mono'",
          "'SF Mono'",
          "'Cascadia Code'",
          "'Fira Code'",
          'monospace',
        ],
      },
      fontSize: {
        xs: ['11px', '16px'],
        sm: ['12px', '18px'],
        base: ['13px', '20px'],
        md: ['14px', '22px'],
        lg: ['16px', '24px'],
        xl: ['18px', '28px'],
      },
    },
  },
  plugins: [],
}
