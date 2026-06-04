/** Tab buttons filtered upstream via usePermTabs. */
export default function PermTabBar({ tabs, active, onChange }) {
  if (!tabs?.length) return null;
  return (
    <div className="tabs">
      {tabs.map((t) => (
        <button
          key={t.id}
          type="button"
          className={`tab ${active === t.id ? "active" : ""}`}
          onClick={() => onChange(t.id)}
        >
          {t.label}
        </button>
      ))}
    </div>
  );
}
