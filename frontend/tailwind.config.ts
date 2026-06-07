import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#1d9bf0',
          hover: '#1a8cd8',
        },
        ink: '#0f1419',
        muted: '#536471',
        line: '#eff3f4',
      },
      maxWidth: {
        feed: '600px',
      },
    },
  },
  plugins: [],
};

export default config;
