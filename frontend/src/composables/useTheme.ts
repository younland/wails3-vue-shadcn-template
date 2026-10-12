import { computed, ref } from "vue";

export type ThemeMode = "light" | "dark" | "system";

const STORAGE_KEY = "app:theme";
const MODES: ThemeMode[] = ["light", "dark", "system"];

function storedMode(): ThemeMode {
  const value = localStorage.getItem(STORAGE_KEY);
  return MODES.includes(value as ThemeMode) ? (value as ThemeMode) : "system";
}

const mode = ref<ThemeMode>(storedMode());
const systemDark = ref(false);
const media = window.matchMedia("(prefers-color-scheme: dark)");

function apply() {
  systemDark.value = media.matches;
  const dark = mode.value === "dark" || (mode.value === "system" && systemDark.value);
  document.documentElement.classList.toggle("dark", dark);
}

media.addEventListener("change", apply);
apply();

const isDark = computed(() => mode.value === "dark" || (mode.value === "system" && systemDark.value));

export function useTheme() {
  function setMode(next: ThemeMode) {
    mode.value = next;
    localStorage.setItem(STORAGE_KEY, next);
    apply();
  }
  return { mode, isDark, setMode };
}
