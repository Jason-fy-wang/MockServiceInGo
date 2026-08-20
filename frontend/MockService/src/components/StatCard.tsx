import { TYPE_CONFIG } from "../constants/mock";

interface StatCardProps {
  label: string;
  count: number;
  active: boolean;
  onClick: () => void;
}

/**
 * StatCard — clickable tab card showing count + label.
 * Used for All / HTTP / SSE / WebSocket filter tabs.
 */
export default function StatCard({
  label,
  count,
  active,
  onClick,
}: StatCardProps) {
  const style = TYPE_CONFIG[label] ?? TYPE_CONFIG.All;

  return (
    <button
      onClick={onClick}
      className={`rounded-xl p-5 text-left transition-all cursor-pointer ${
        active
          ? `${style.bg} ring-2 ring-offset-2 ring-blue-400 shadow-md`
          : "shadow-sm hover:shadow"
      }`}
    >
      <div
        className={`${active ? `${style.text} opacity-100` : `${style.textInactive} opacity-80`}`}
      >
        {count}
      </div>
      <div
        className={`${active ? `${style.text} opacity-100` : `${style.textInactive} opacity-80`}`}
      >
        {label}
      </div>
    </button>
  );
}
