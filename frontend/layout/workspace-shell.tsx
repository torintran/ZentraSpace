'use client';

import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import { useEffect, useState, type ReactNode } from 'react';
import { SiteHeader } from './site-header';

const sidebarPreference = 'zentraspace.sidebar.collapsed';

export function WorkspaceShell({
  brand,
  account,
  sidebar,
  children,
}: {
  brand: ReactNode;
  account: ReactNode;
  sidebar: ReactNode;
  children: ReactNode;
}) {
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
    <div className="workspace-shell" data-sidebar-collapsed={collapsed}>
      <SiteHeader
        brand={brand}
        account={account}
        control={
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
          </button>
        }
      />
      <div className="workspace-frame with-sidebar">
        {sidebar}
        {children}
      </div>
    </div>
  );
}
