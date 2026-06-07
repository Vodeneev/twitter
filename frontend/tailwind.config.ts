import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#FC3F1D',
          hover: '#E03518',
        },
        ink: '#21201F',
        muted: '#6D6D6D',
        line: '#F0F0F0',
      },
      maxWidth: {
        feed: '600px',
      },
    },
  },
  plugins: [],
};

export default config;
