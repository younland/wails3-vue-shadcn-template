// 界面提示统一用应用内 toast（vue-sonner）。
// 主窗口隐藏在托盘时，后端会改用系统通知（见 notifyUser），否则窗口隐藏就看不到提示。
import { toast } from "vue-sonner";

/** 操作成功提示，例如 notifySuccess("已连接", "「测试网盘」已挂载") */
export function notifySuccess(title: string, body = "") {
  toast.success(title, { description: body || undefined });
}

/** 操作失败提示：title 说明失败的动作，message 是具体原因 */
export function notifyError(message: string, title = "操作失败") {
  toast.error(title, { description: message || undefined });
}

/** 后端事件（app:toast）转发的提示 */
export function notifyFromBackend(level: string, title: string, body: string) {
  if (level === "error") {
    notifyError(body, title);
    return;
  }
  notifySuccess(title, body);
}
