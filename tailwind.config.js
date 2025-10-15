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
        'text-primary': '#111827',
        'text-secondary': '#6b7280',
        'accent': {
          'DEFAULT': '#ea580c',
          'hover': '#f9fafb',
        },
        'high-sec': '#10B981',
        'low-sec': '#F59E0B',
        'null-sec': '#EF4444',
      },
      borderRadius: {
        'DEFAULT': '0.5rem',
      },
      fontFamily: {
        sans: ['Montserrat', 'sans-serif'],
      },
      keyframes: {
        fadeIn: {
          'from': { opacity: '0', transform: 'translateY(5px)' },
          'to': { opacity: '1', transform: 'translateY(0)' },
        }
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-out forwards',
      }
    },
  },
  plugins: [
    plugin(function({ addBase, addComponents, addUtilities, theme }) {
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
          '&:hover': { textDecoration: 'underline' }
        },
      });
      
      // --- Reusable Component Classes (Fully Populated) ---
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
            '@apply text-red-600 font-semibold': {},
        },
        '.nav-icon': {
            '@apply w-5 h-5 mr-3 flex-shrink-0': {},
        },
        '.btn': {
            '@apply inline-flex items-center justify-center px-4 py-2 font-semibold rounded-lg shadow-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2': {},
        },
        '.btn-primary': {
            '@apply btn bg-orange-600 text-white hover:bg-orange-700 focus:ring-orange-500': {},
        },
        '.btn-secondary': {
            '@apply btn bg-gray-200 text-gray-800 hover:bg-gray-300 focus:ring-gray-400': {},
        },
      });

      // --- Custom Utility Classes ---
      addUtilities({
        '.animate-fade-in': {
          animation: theme('animation.fade-in'),
        }
      });
    })
  ],
}