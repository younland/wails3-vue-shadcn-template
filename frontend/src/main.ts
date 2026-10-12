import { createApp } from "vue";
import App from "./App.vue";
// vue-sonner 的定位与布局样式，缺少它 toast 不会显示
import "vue-sonner/style.css";
import "./style.css";

createApp(App).mount("#app");
