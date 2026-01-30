/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Anime-inspired sakura palette
        sakura: {
          50: '#FFF5F7',
          100: '#FFE3E9',
          200: '#FFC7D3',
          300: '#FFABC0',
          400: '#FF8FAC',
          500: '#FF69B4',  // Hot Pink (SakuraPink)
          600: '#E54B96',
          700: '#CC3378',
          800: '#B21B5A',
          900: '#991145',
        },
        electric: {
          50: '#E6F9FF',
          100: '#CCF3FF',
          200: '#99E7FF',
          300: '#66DBFF',
          400: '#33CFFF',
          500: '#00D9FF',  // Bright Cyan (ElectricBlue)
          600: '#00B8D9',
          700: '#0097B3',
          800: '#00768C',
          900: '#005566',
        },
        neon: {
          50: '#F5F0FF',
          100: '#EBE0FF',
          200: '#D7C1FF',
          300: '#C3A2FF',
          400: '#BD93F9',  // Neon Purple
          500: '#A374E8',
          600: '#8955D7',
          700: '#6F36C6',
          800: '#5517B5',
          900: '#3B00A4',
        },
        mint: {
          50: '#EFFFEF',
          100: '#D6FFD9',
          200: '#ADFFB3',
          300: '#84FF8D',
          400: '#5BFA67',
          500: '#50FA7B',  // Mint Green
          600: '#32E14F',
          700: '#14C823',
          800: '#00AF00',
          900: '#009600',
        },
        sunset: {
          50: '#FFF7ED',
          100: '#FFEDDB',
          200: '#FFD9B7',
          300: '#FFC593',
          400: '#FFB86C',  // Sunset Orange
          500: '#FFA548',
          600: '#FF9224',
          700: '#FF7F00',
          800: '#DB6C00',
          900: '#B75900',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
        mono: ['Fira Code', 'Menlo', 'Monaco', 'Courier New', 'monospace'],
      },
      animation: {
        'sakura-float': 'float 6s ease-in-out infinite',
        'pulse-glow': 'pulse-glow 2s ease-in-out infinite',
      },
      keyframes: {
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-20px)' },
        },
        'pulse-glow': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
      },
    },
  },
  plugins: [],
}
