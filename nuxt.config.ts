// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ["@nuxtjs/tailwindcss", "shadcn-nuxt", '@nuxtjs/color-mode'],
  colorMode: {
    preference: 'dark'
  },
  shadcn: {
    componentDir: './components/ui'
  }
})