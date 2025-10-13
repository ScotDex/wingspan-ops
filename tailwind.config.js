// tailwind.config.js

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    // Add paths to all of your template files here
  "./templates/**/*.html",
  ],
  theme: {
    extend: {
      // This is where we define the color palette from your CSS variables
      colors: {
        'bg': {
          'color': '#ffffff',
        },
        'panel': {
          'bg': '#ffffff',
        },
        'border': {
          'color': '#e5e7eb',
        },
        'text': {
          'primary': '#1f2937',
          'secondary': '#6b7280',
        },
        'accent': {
          'orange': '#ea580c',
          'hover': '#f3f4f6',
        },
      },

      
      // This defines the default border radius from your CSS variable
      borderRadius: {
        'DEFAULT': '0.5rem', // Corresponds to --radius
      },
      fontFamily: {
        sans: ['Montserrat', 'sans-serif'],
      },
    },
  },
  // This is where we load our custom plugin
  plugins: [
    require('./wingspan-theme.js'),
  ],
}