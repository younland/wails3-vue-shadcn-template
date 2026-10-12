<script setup lang="ts">
import { onMounted, ref } from "vue";
import { Moon, Sun } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Toaster } from "@/components/ui/sonner";
import { useTheme } from "./composables/useTheme";
import { notifySuccess } from "./lib/notify";

// Wails 生成的 TypeScript 绑定放在 frontend/bindings/<Go 模块名>/ 目录下，
// 模块名随 go.mod 变化，这里按目录通配引入，避免在源码里写死模块名。
type GreetService = {
  Greet(name: string): Promise<string>;
};

const greetModules = import.meta.glob<GreetService>("../bindings/*/greetservice");
const loadGreetService = Object.values(greetModules)[0];

const name = ref("");
const result = ref("");
const ready = ref(false);
const { isDark, setMode } = useTheme();

onMounted(async () => {
  if (!loadGreetService) {
    result.value = "未找到 Wails 绑定，请先运行 `wails3 generate bindings`";
    return;
  }
  ready.value = true;
});

async function greet() {
  if (!loadGreetService) return;
  const service = await loadGreetService();
  result.value = await service.Greet(name.value);
  notifySuccess("Greet 调用成功", result.value);
}
</script>

<template>
  <div class="bg-background flex h-screen items-center justify-center p-6">
    <Card class="w-full max-w-md">
      <CardHeader>
        <div class="flex items-center justify-between">
          <div>
            <CardTitle>{{ result ? "Hello" : "Welcome" }}</CardTitle>
            <CardDescription>Wails v3 + Vue 3 + shadcn-vue 模板演示</CardDescription>
          </div>
          <Button variant="ghost" size="icon" aria-label="切换主题" @click="setMode(isDark ? 'light' : 'dark')">
            <Sun v-if="isDark" />
            <Moon v-else />
          </Button>
        </div>
      </CardHeader>
      <CardContent class="flex flex-col gap-3">
        <form class="flex gap-2" @submit.prevent="greet">
          <Input v-model="name" placeholder="输入你的名字…" :disabled="!ready" />
          <Button type="submit" :disabled="!ready">Greet</Button>
        </form>
        <p v-if="result" class="text-muted-foreground text-sm">{{ result }}</p>
      </CardContent>
    </Card>

    <Toaster :theme="isDark ? 'dark' : 'light'" position="top-center" />
  </div>
</template>
