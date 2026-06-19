import { useState, useCallback } from "react";
import type { ToastMessage, ToastConfig } from "./Toast";

/**
 * useToast — hook that returns [toasts, toast] where:
 * - toasts: current ToastMessage[] for rendering <ToastStack />
 * - toast(config): function to push a new notification
 */
export function useToast() {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toast = useCallback(
    ({ variant, text, duration = 3000 }: ToastConfig) => {
      const id = Date.now();
      setToasts((prev) => [...prev, { id, variant, text }]);
      if (duration > 0) {
        setTimeout(() => remove(id), duration);
      }
    },
    [remove],
  );

  return { toasts, toast } as const;
}
