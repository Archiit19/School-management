import { useState, useEffect, useMemo } from "react";
import { useAuth } from "../context/AuthContext";

/**
 * Filter tabs by permission and keep the active tab valid.
 * @param {{ id: string, label: string, perm?: string, any?: string[] }[]} tabDefs
 * @param {string} [defaultId]
 */
export function usePermTabs(tabDefs, defaultId) {
  const { hasPerm } = useAuth();

  const visibleTabs = useMemo(
    () =>
      tabDefs.filter((t) => {
        if (t.perm) return hasPerm(t.perm);
        if (t.any?.length) return t.any.some((p) => hasPerm(p));
        return true;
      }),
    [tabDefs, hasPerm],
  );

  const initial = defaultId && visibleTabs.some((t) => t.id === defaultId)
    ? defaultId
    : visibleTabs[0]?.id ?? "";

  const [tab, setTab] = useState(initial);

  useEffect(() => {
    if (visibleTabs.length === 0) return;
    if (!visibleTabs.some((t) => t.id === tab)) {
      setTab(visibleTabs[0].id);
    }
  }, [visibleTabs, tab]);

  return { visibleTabs, tab, setTab };
}
