/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./ui/**/*.{html,js}"],
  theme: {
    extend: {
colors:{
      base: 'rgb(25, 23, 36)',
      gold: 'rgb(246, 193, 119)',
      muted: 'rgb(110, 106, 134)',
      text: 'rgb(224, 222, 244)',
      rose: 'rgb(235, 188, 186)',
      foam: 'rgb(156, 207, 216)',
      highlight :{
        low: 'rgb(33, 32, 46)',
        med: 'rgb(64, 61, 82)',
        high: 'rgb(82, 79, 103)',
      },
      love:'rgb(235, 111, 146)',
      iris:'rgb(196, 167, 231)',

    },
    },
  },
  plugins: [],
}

