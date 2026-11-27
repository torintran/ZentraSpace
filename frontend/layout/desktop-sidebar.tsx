'use client';

import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import { useEffect, useState, type ReactNode } from 'react';

const sidebarPreference = 'zentraspace.sidebar.collapsed';

export function DesktopSidebar({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    try {
      setCollapsed(localStorage.getItem(sidebarPreference) === 'true');
    } catch {
      // The menu still works when the browser blocks storage.
    }
  }, []);

  function toggle() {
    const next = !collapsed;
    setCollapsed(next);
    try {
      localStorage.setItem(sidebarPreference, String(next));
    } catch {
      // Keep the current view even if the preference cannot be saved.
    }
  }

  const label = collapsed ? 'Mở rộng thanh bên' : 'Thu gọn thanh bên';
  const Icon = collapsed ? PanelLeftOpen : PanelLeftClose;
  return (
    <div className="workspace-sidebar-desktop" data-collapsed={collapsed}>
      <button
        type="button"
        className="workspace-sidebar-toggle"
        aria-label={label}
        aria-expanded={!collapsed}
        aria-controls="workspace-desktop-links"
        title={label}
        onClick={toggle}
      >
        <Icon size={20} aria-hidden />
        <span className="workspace-link-label">Thu gọn</span>
      </button>
      <div id="workspace-desktop-links">{children}</div>
    </div>
  );
}
