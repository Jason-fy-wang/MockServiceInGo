export type ToastVariant = "success" | "error" | "warning" | "info";

export interface ToastMessage {
  id: number;
  variant: ToastVariant;
  text: string;
}

export interface ToastConfig {
  variant: ToastVariant;
  text: string;
  duration?: number; // ms, default 3000
}

const VARIANT_STYLES: Record<
  ToastVariant,
  { bg: string; border: string; icon: string; text: string }
> = {
  success: {
    bg: "bg-green-50",
    border: "border-green-200",
    icon: "text-green-500",
    text: "text-green-800",
  },
  error: {
    bg: "bg-red-50",
    border: "border-red-200",
    icon: "text-red-500",
    text: "text-red-800",
  },
  warning: {
    bg: "bg-yellow-50",
    border: "border-yellow-200",
    icon: "text-yellow-600",
    text: "text-yellow-800",
  },
  info: {
    bg: "bg-blue-50",
    border: "border-blue-200",
    icon: "text-blue-500",
    text: "text-blue-800",
  },
};

const ICONS: Record<ToastVariant, React.ReactNode> = {
  success: (
    <svg
      className="w-4 h-4"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M5 13l4 4L19 7"
      />
    </svg>
  ),
  error: (
    <svg
      className="w-4 h-4"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M6 18L18 6M6 6l12 12"
      />
    </svg>
  ),
  warning: (
    <svg
      className="w-4 h-4"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z"
      />
    </svg>
  ),
  info: (
    <svg
      className="w-4 h-4"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  ),
};

/* ───────── Render component ───────── */
interface ToastStackProps {
  toasts: ToastMessage[];
  onClose: (id: number) => void;
}

/**
 * ToastStack — renders a vertical stack of toast notifications.
 * Place once at the top of your layout (e.g. inside App root div).
 */
export default function ToastStack({ toasts, onClose }: ToastStackProps) {
  if (toasts.length === 0) return null;

  return (
    <div className="fixed top-6 right-6 z-[100] flex flex-col gap-2 w-full max-w-sm pointer-events-none">
      {toasts.map((t) => {
        const style = VARIANT_STYLES[t.variant];
        return (
          <div
            key={t.id}
            role="alert"
            className={`pointer-events-auto flex items-start gap-3 px-4 py-3 rounded-xl border shadow-lg animate-slide-in ${style.bg} ${style.border}`}
          >
            <span className={`mt-0.5 flex-shrink-0 ${style.icon}`}>
              {ICONS[t.variant]}
            </span>
            <p
              className={`flex-1 text-sm font-medium leading-snug ${style.text}`}
            >
              {t.text}
            </p>
            <button
              onClick={() => onClose(t.id)}
              className={`flex-shrink-0 p-0.5 rounded hover:bg-black/5 cursor-pointer transition-colors ${style.icon}`}
            >
              <svg
                className="w-3.5 h-3.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
        );
      })}
    </div>
  );
}
