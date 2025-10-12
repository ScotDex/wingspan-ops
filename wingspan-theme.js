// wingspan-theme.js

const plugin = require('tailwindcss/plugin');

module.exports = plugin(function({ addBase, addComponents, theme }) {
    // Add base styles for html, body, etc.
    addBase({
        'html': {
            // Fluid Typography
            fontSize: 'clamp(15px, 0.9vw + 12px, 18px)',
            // Native UI Enhancements
            colorScheme: 'light',
            accentColor: theme('colors.accent.orange'),
            '@media (prefers-reduced-motion: no-preference)': {
                scrollBehavior: 'smooth',
            },
        },
        'html, body': {
            height: '100%',
            margin: '0',
            padding: '0',
            fontFamily: '"Inter", sans-serif',
            '-webkit-font-smoothing': 'antialiased',
            color: theme('colors.text.primary'),
            backgroundColor: theme('colors.bg.color'),
        },
        'body': {
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
        },
        // Accessibility and Reduced Motion
        '@media (prefers-reduced-motion: reduce)': {
            '*': {
                animation: 'none !important',
                transition: 'none !important',
                scrollBehavior: 'auto !important',
            },
        },
        ':where(a, button, [role="button"], input, select, textarea, summary, [tabindex]:not([tabindex="-1"])):focus-visible': {
            outline: `2px solid ${theme('colors.accent.orange')}`,
            outlineOffset: '2px',
            borderRadius: '0.25rem',
        },
        // Table styles
        'table': {
            width: '100%',
            borderCollapse: 'collapse',
        },
        'th, td': {
            padding: '12px 15px',
            textAlign: 'left',
            borderBottom: `1px solid ${theme('colors.border.color')}`,
            whiteSpace: 'nowrap',
        },
        'th': {
            color: theme('colors.text.secondary'),
            fontSize: '0.8em',
            fontWeight: 'bold',
            textTransform: 'uppercase',
            letterSpacing: '1px',
            borderBottomWidth: '2px',
        },
        'tbody tr:nth-child(even)': {
            backgroundColor: theme('colors.accent.hover'),
        },
        'tbody tr:hover': {
            backgroundColor: '#e5e7eb', // Kept specific color as it differs from theme hover
        },
    });

    // Add component classes
    addComponents({
        // Core Layout Structure
        '.container': {
            width: '95vw',
            height: '95vh',
            overflow: 'hidden',
            display: 'flex',
            border: `1px solid ${theme('colors.border.color')}`,
            borderRadius: theme('borderRadius.DEFAULT'),
            boxShadow: '0 8px 20px -12px rgb(0 0 0 / 0.18), 0 4px 10px -6px rgb(0 0 0 / 0.10)',
        },
        '.sidebar': {
            width: '14rem',
            backgroundColor: '#f9fafb', // Specific color from original CSS
            padding: '1rem',
            borderRight: `1px solid ${theme('colors.border.color')}`,
            display: 'flex',
            flexDirection: 'column',
        },
        '.main-view': {
            flex: '1',
            display: 'flex',
            flexDirection: 'column',
        },
        '.main-header': {
            padding: '1rem',
            borderBottom: `1px solid ${theme('colors.border.color')}`,
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
        },
        '.main-content': {
            flex: '1',
            padding: '1.5rem',
            overflowY: 'auto',
            display: 'grid',
            gridTemplateColumns: '1fr',
            gap: 'clamp(0.75rem, 1.2vw, 1.5rem)',
            '@screen md': {
                gridTemplateColumns: 'repeat(3, 1fr)',
            },
        },
        '.connections-panel': {
            '@screen md': {
                gridColumn: 'span 2 / span 2',
            },
        },

        // Component Styles
        '.glass-panel': {
            background: theme('colors.panel.bg'),
            border: `1px solid ${theme('colors.border.color')}`,
            borderRadius: theme('borderRadius.DEFAULT'),
            overflowX: 'auto',
        },
        '.sidebar-header': {
            marginBottom: '2rem',
        },
        '.eve-header, .section-header': {
            fontWeight: '500',
            color: theme('colors.accent.orange'),
            textTransform: 'uppercase',
            letterSpacing: '1.5px',
            fontSize: '1.125rem',
            marginBottom: '1rem',
            paddingLeft: '0.5rem',
        },
        '.eve-accent-border': {
            borderLeft: `3px solid ${theme('colors.accent.orange')}`,
        },
        '.sidebar-nav a.nav-link': {
            display: 'flex',
            alignItems: 'center',
            padding: '0.5rem',
            borderRadius: '0.375rem',
            textDecoration: 'none',
            color: theme('colors.text.primary'),
            transition: 'background-color 0.2s',
            '&:hover': {
                backgroundColor: theme('colors.accent.hover'),
            },
            '&.active': {
                backgroundColor: theme('colors.accent.hover'),
                color: theme('colors.accent.orange'),
                fontWeight: '600',
            },
        },
        '.nav-icon': {
            width: '1.25rem',
            height: '1.25rem',
            marginRight: '0.75rem',
            flexShrink: '0',
        },
        '.sidebar-footer': {
            marginTop: 'auto',
            fontSize: '0.75rem',
            color: theme('colors.text.secondary'),
            '& a': {
                color: theme('colors.text.secondary'),
                textDecoration: 'none',
                transition: 'color 0.2s',
                '&:hover': {
                    color: theme('colors.accent.orange'),
                },
            },
        },
        '.site-footer': {
            position: 'fixed',
            bottom: '0',
            left: '0',
            width: '100%',
            backgroundColor: theme('colors.accent.orange'),
            padding: '0.5rem',
            textAlign: 'center',
            zIndex: '100',
            '& p': {
                margin: '0',
                fontSize: '0.75rem',
                color: '#ffffff',
            },
        },
        
        // Alpha Badge
        'body::after': {
            content: "'ALPHA'",
            position: 'fixed',
            top: '12px',
            right: '12px',
            zIndex: '9999',
            backgroundColor: theme('colors.accent.orange'),
            color: '#fff',
            padding: '4px 8px',
            borderRadius: '4px',
            fontSize: '12px',
            fontWeight: 'bold',
            letterSpacing: '1px',
        },

        // Table specific helpers
        '.status-critical': {
            color: '#dc2626',
            fontWeight: '600',
        },
        'th[data-sortable]': {
            cursor: 'pointer',
            userSelect: 'none',
            '&:hover': {
                color: theme('colors.text.primary'),
            },
        },
        '.sort-indicator': {
            display: 'inline-block',
            width: '0.8em',
            height: '0.8em',
            marginLeft: '0.5em',
            verticalAlign: 'middle',
            opacity: '0.4',
        },
        'th.sort-asc .sort-indicator, th.sort-desc .sort-indicator': {
            opacity: '1',
        },
        'th.sort-asc .sort-indicator::before': {
            content: "'▼'",
        },
        'th.sort-desc .sort-indicator::before': {
            content: "'▲'",
        },
        
        // Leaderboard Styles
        '.leaderboard-list': {
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
            fontSize: '0.875rem',
        },
        '.leaderboard-item': {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            backgroundColor: '#f9fafb',
            padding: '0.5rem 0.75rem',
            borderRadius: '0.25rem',
        },
        '.scan-count': {
            fontWeight: '700',
            color: theme('colors.accent.orange'),
        },
        '.no-data-message': {
            fontSize: '0.75rem',
            color: theme('colors.text.secondary'),
        },

        // Route Planner Result Styles
        '.results-container': {
            marginTop: '2rem',
            borderTop: `1px solid ${theme('colors.border.color')}`,
            paddingTop: '1.5rem',
        },
        '.route-steps': {
            listStyle: 'none',
            padding: '0',
            marginTop: '1rem',
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
        },
        '.route-step': {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            backgroundColor: theme('colors.accent.hover'),
            padding: '0.75rem 1rem',
            borderRadius: theme('borderRadius.DEFAULT'),
            border: `1px solid ${theme('colors.border.color')}`,
        },
        '.step-info': {
            display: 'flex',
            alignItems: 'center',
            gap: '1rem',
        },
        '.step-number': {
            color: theme('colors.text.secondary'),
            fontWeight: '500',
        },
        '.step-system': {
            fontWeight: '600',
            fontSize: '1.1rem',
            '& span': {
                fontSize: '0.85rem',
                fontWeight: '400',
                color: theme('colors.text.secondary'),
            },
        },
        '.jump-type': {
            fontSize: '0.75rem',
            fontWeight: 'bold',
            textTransform: 'uppercase',
            padding: '0.25rem 0.6rem',
            borderRadius: '999px',
            color: '#fff',
        },
        '.jump-type-wormhole': {
            backgroundColor: theme('colors.accent.orange'),
        },
        '.jump-type-stargate': {
            backgroundColor: theme('colors.text.secondary'),
        },
        '.high-sec': { color: '#10B981' },
        '.low-sec': { color: '#F59E0B' },
        '.null-sec': { color: '#EF4444' },
    });
});