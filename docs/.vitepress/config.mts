import { defineConfig } from "vitepress"

const repo = "https://github.com/bzdvdn/speckeep"

const enSidebar = [
  {
    text: "Guide",
    items: [
      { text: "Overview", link: "/en/overview" },
      { text: "Workflow Model", link: "/en/workflow" },
      { text: "CLI Reference", link: "/en/cli" },
      { text: "Examples & Recipes", link: "/en/examples" },
      { text: "FAQ", link: "/en/faq" },
    ],
  },
  {
    text: "Reference",
    items: [
      { text: "Architecture", link: "/en/architecture" },
      { text: "Agents", link: "/en/agents" },
      { text: "Language & Config", link: "/en/language-and-config" },
      { text: "Glossary", link: "/en/glossary" },
      { text: "Roadmap", link: "/en/roadmap" },
      { text: "Self-Hosting", link: "/en/self-hosting" },
    ],
  },
]

const ruSidebar = [
  {
    text: "Руководство",
    items: [
      { text: "Обзор", link: "/ru/overview" },
      { text: "Модель workflow", link: "/ru/workflow" },
      { text: "Справочник CLI", link: "/ru/cli" },
      { text: "Примеры и рецепты", link: "/ru/examples" },
      { text: "FAQ", link: "/ru/faq" },
    ],
  },
  {
    text: "Справочник",
    items: [
      { text: "Архитектура", link: "/ru/architecture" },
      { text: "Агенты", link: "/ru/agents" },
      { text: "Язык и конфигурация", link: "/ru/language-and-config" },
      { text: "Глоссарий", link: "/ru/glossary" },
      { text: "Roadmap", link: "/ru/roadmap" },
      { text: "Self-hosting", link: "/ru/self-hosting" },
    ],
  },
]

export default defineConfig({
  base: "/speckeep/",
  title: "speckeep",
  description: "Strict, lightweight spec-driven development for coding agents",
  lastUpdated: true,
  sitemap: { hostname: "https://bzdvdn.github.io/speckeep/" },
  head: [["meta", { name: "theme-color", content: "#3c8772" }]],
  themeConfig: {
    socialLinks: [{ icon: "github", link: repo }],
    search: { provider: "local" },
    editLink: {
      pattern: `${repo}/edit/master/docs/:path`,
      text: "Edit this page on GitHub",
    },
    footer: {
      message: "Released under the MIT License.",
      copyright: "Copyright © 2026 speckeep",
    },
  },
  locales: {
    en: {
      label: "English",
      lang: "en",
      link: "/en/",
      themeConfig: {
        nav: [
          { text: "Guide", link: "/en/overview" },
          { text: "CLI", link: "/en/cli" },
          { text: "Examples", link: "/en/examples" },
          { text: "Agents", link: "/en/agents" },
          { text: "GitHub", link: repo },
        ],
        sidebar: enSidebar,
        outline: { label: "On this page" },
        editLink: { pattern: `${repo}/edit/master/docs/:path`, text: "Edit this page on GitHub" },
      },
    },
    ru: {
      label: "Русский",
      lang: "ru",
      link: "/ru/",
      themeConfig: {
        nav: [
          { text: "Руководство", link: "/ru/overview" },
          { text: "CLI", link: "/ru/cli" },
          { text: "Примеры", link: "/ru/examples" },
          { text: "Агенты", link: "/ru/agents" },
          { text: "GitHub", link: repo },
        ],
        sidebar: ruSidebar,
        outline: { label: "На этой странице" },
        editLink: { pattern: `${repo}/edit/master/docs/:path`, text: "Изменить страницу на GitHub" },
      },
    },
  },
})
