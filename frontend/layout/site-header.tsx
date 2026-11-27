import type { ReactNode } from 'react';

export function SiteHeader({
  brand,
  account,
  control,
}: {
  brand: ReactNode;
  account: ReactNode;
  control?: ReactNode;
}) {
  return (
    <header className={control ? 'bar workspace-header' : 'bar'}>
      <div className="bar-inner">
        <div className="bar-start">
          {control}
          {brand}
        </div>
        {account}
      </div>
    </header>
  );
}
