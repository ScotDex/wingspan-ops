// tailwind.config.js

const plugin = require('tailwindcss/plugin');

module.exports = {
  content: [
    "./templates/**/*.html",
  ],
  theme: {
    extend: {
      colors: {
        'background': '#ffffff',
        'panel': '#ffffff',
        'border': '#e5e7eb',
        'text-primary': '#111827', // Dark gray for readability
        'text-secondary': '#6b7280',
        'accent': {
          'DEFAULT': '#ea580c', // Default orange
          'hover': '#f9fafb',   // Light gray for hover states
        },
        'high-sec': '#10B981',
        'low-sec': '#F59E0B',
        'null-sec': '#EF4444',
      },
      borderRadius: {
        'DEFAULT': '0.5rem',
      },
      fontFamily: {
        // Montserrat is now the default sans-serif font for the project
        sans: ['Montserrat', 'sans-serif'],
      },
    },
  },
  plugins: [
    plugin(function({ addBase, addComponents, theme }) {
      // --- Base Styles ---
      addBase({
        'html': {
          fontFamily: theme('fontFamily.sans'),
          color: theme('colors.text-primary'),
          backgroundColor: theme('colors.background'),
          '-webkit-font-smoothing': 'antialiased',
        },
        'a': {
          color: theme('colors.accent.DEFAULT'),
          textDecoration: 'none',
          '&:hover': {
            textDecoration: 'underline',
          }
        },
        'table': {
            width: '100%',
            borderCollapse: 'collapse',
        },
        'th, td': {
            padding: '12px 15px',
            textAlign: 'left',
            borderBottom: `1px solid ${theme('colors.border')}`,
        },
      });
      
      // --- Reusable Component Classes ---
      addComponents({
        '.container': {
          width: '95vw',
          height: '95vh',
          overflow: 'hidden',
          display: 'flex',
          border: `1px solid ${theme('colors.border')}`,
          borderRadius: theme('borderRadius.DEFAULT'),
          boxShadow: '0 8px 20px -12px rgb(0 0 0 / 0.18), 0 4px 10px -6px rgb(0 0 0 / 0.10)',
        },
        '.sidebar': {
          width: '14rem',
          backgroundColor: theme('colors.accent.hover'),
          padding: '1rem',
          borderRight: `1px solid ${theme('colors.border')}`,
        },
        '.main-view': {
          flex: '1',
          display: 'flex',
          flexDirection: 'column',
        },
        '.glass-panel': {
          background: theme('colors.panel'),
          border: `1px solid ${theme('colors.border')}`,
          borderRadius: theme('borderRadius.DEFAULT'),
          overflowX: 'auto',
        },
        '.section-header': {
            fontWeight: '500',
            color: theme('colors.accent.DEFAULT'),
            textTransform: 'uppercase',
            letterSpacing: '1.5px',
            fontSize: '1.125rem',
            marginBottom: '1rem',
            paddingLeft: '0.5rem',
            borderLeft: `3px solid ${theme('colors.accent.DEFAULT')}`,
        },
        '.status-critical': {
            color: '#dc2626',
            fontWeight: '600',
        },
      });
    })
  ],
}