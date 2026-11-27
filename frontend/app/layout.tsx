import '@mantine/core/styles.css';
import 'leaflet/dist/leaflet.css';
import './screen.css';
import { mantineHtmlProps } from '@mantine/core';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { Providers } from './providers';

export const metadata: Metadata = {
  title: 'ZentraSpace',
  icons: '/assets/icon.svg',
  robots: { index: false, follow: false },
};

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="vi" {...mantineHtmlProps} data-mantine-color-scheme="light">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
